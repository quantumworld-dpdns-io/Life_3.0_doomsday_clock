import { SCENARIOS, type ScenarioId } from "./scenarios";
import type { ClockState, ScenarioSignal } from "./types";

const MOCK_TITLES: Record<ScenarioId, string> = {
  LIBERTARIAN_UTOPIA: "Open compute compact expands cross-border AI access",
  BENEVOLENT_DICTATOR: "National planning office delegates triage to foundation model",
  EGALITARIAN_UTOPIA: "AI dividend pilot links productivity gains to public services",
  GATEKEEPER: "Safety institute proposes capability gate for frontier training",
  PROTECTOR_GOD: "Autonomous mediation system defuses regional cyber escalation",
  ENSLAVED_GOD: "Closed lab reports major capability jump under sovereign contract",
  CONQUERORS: "Autonomous targeting stack enters active conflict zone",
  DESCENDANTS: "Neural interface trial reports sustained synthetic memory handoff",
  ZOOKEEPER: "Algorithmic welfare controller overrides municipal budget votes",
  NINETEEN_EIGHTY_FOUR: "Real-time biometric network expands to transit corridors",
  REVERT: "Coordinated sabotage hits AI datacenter backbone",
  SELF_DESTRUCTION: "Oversight board warns of loss-of-control bioautomation pathway",
};

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

export const makeMockClockState = (date = new Date()): ClockState => {
  const pulse = Math.sin(date.getTime() / 42_000);
  const slowPulse = Math.cos(date.getTime() / 91_000);
  const weights = SCENARIOS.map((scenario, index) => {
    const wave = Math.sin(date.getTime() / (28_000 + index * 2_750) + index);
    const base = 0.08 + scenario.severity * 0.34;
    return clamp(base + wave * 0.045 + (index % 3) * 0.012, 0.03, 0.96);
  });

  weights[9] = clamp(weights[9] + 0.1 + pulse * 0.03, 0, 1);
  weights[11] = clamp(weights[11] + 0.14 + slowPulse * 0.04, 0, 1);

  const dominantIndex = weights.reduce(
    (best, weight, index) => (weight > weights[best] ? index : best),
    0,
  );
  const aggregateRisk =
    weights.reduce((sum, weight, index) => sum + weight * SCENARIOS[index].severity, 0) /
    SCENARIOS.length;

  return {
    minutesToMidnight: clamp(60 * (1 - aggregateRisk * 1.85), 0.6, 58),
    dominantScenario: SCENARIOS[dominantIndex].id,
    scenarioWeights: weights,
    computedAt: date.toISOString(),
  };
};

export const makeMockSignals = (date = new Date(), limit = 18): ScenarioSignal[] =>
  Array.from({ length: limit }, (_, index) => {
    const scenario = SCENARIOS[(index * 5 + date.getUTCMinutes()) % SCENARIOS.length];
    const confidence = clamp(0.42 + scenario.severity * 0.4 + Math.sin(index + date.getTime() / 60_000) * 0.08, 0.2, 0.98);
    return {
      scenario: scenario.id,
      confidence,
      sourceUrl: `mock://life3/signals/${scenario.id.toLowerCase()}/${index + 1}`,
      timestamp: new Date(date.getTime() - index * 17 * 60_000).toISOString(),
      title: MOCK_TITLES[scenario.id],
    };
  });
