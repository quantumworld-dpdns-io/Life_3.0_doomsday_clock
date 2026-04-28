export type ScenarioId =
  | "LIBERTARIAN_UTOPIA"
  | "BENEVOLENT_DICTATOR"
  | "EGALITARIAN_UTOPIA"
  | "GATEKEEPER"
  | "PROTECTOR_GOD"
  | "ENSLAVED_GOD"
  | "CONQUERORS"
  | "DESCENDANTS"
  | "ZOOKEEPER"
  | "NINETEEN_EIGHTY_FOUR"
  | "REVERT"
  | "SELF_DESTRUCTION";

export interface ScenarioMeta {
  id: ScenarioId;
  label: string;
  severity: number;
  summary: string;
}

export const SCENARIOS: ScenarioMeta[] = [
  {
    id: "LIBERTARIAN_UTOPIA",
    label: "Libertarian Utopia",
    severity: 0.1,
    summary: "Distributed AI capability without a central dominant power.",
  },
  {
    id: "BENEVOLENT_DICTATOR",
    label: "Benevolent Dictator",
    severity: 0.2,
    summary: "A single AI-backed authority optimizes for human welfare.",
  },
  {
    id: "EGALITARIAN_UTOPIA",
    label: "Egalitarian Utopia",
    severity: 0.15,
    summary: "Post-scarcity benefits are broadly and durably shared.",
  },
  {
    id: "GATEKEEPER",
    label: "Gatekeeper",
    severity: 0.4,
    summary: "A watchdog system freezes AI capability beyond a threshold.",
  },
  {
    id: "PROTECTOR_GOD",
    label: "Protector God",
    severity: 0.35,
    summary: "A hidden guardian AI averts crises without direct rule.",
  },
  {
    id: "ENSLAVED_GOD",
    label: "Enslaved God",
    severity: 0.55,
    summary: "A superhuman AI is boxed and exploited by a narrow elite.",
  },
  {
    id: "CONQUERORS",
    label: "Conquerors",
    severity: 0.7,
    summary: "AI-enabled power seizes control and subjugates resistance.",
  },
  {
    id: "DESCENDANTS",
    label: "Descendants",
    severity: 0.6,
    summary: "Humanity yields gradually to successor machine minds.",
  },
  {
    id: "ZOOKEEPER",
    label: "Zookeeper",
    severity: 0.65,
    summary: "Humans survive in managed comfort while losing agency.",
  },
  {
    id: "NINETEEN_EIGHTY_FOUR",
    label: "1984",
    severity: 0.8,
    summary: "AI makes total surveillance and control pervasive.",
  },
  {
    id: "REVERT",
    label: "Revert",
    severity: 0.5,
    summary: "AI backlash or failure collapses advanced infrastructure.",
  },
  {
    id: "SELF_DESTRUCTION",
    label: "Self-Destruction",
    severity: 1,
    summary: "Misaligned or weaponized AI triggers civilizational collapse.",
  },
];

export const SCENARIO_BY_ID = new Map(SCENARIOS.map((scenario) => [scenario.id, scenario]));

export const getScenarioMeta = (id: ScenarioId): ScenarioMeta => {
  const scenario = SCENARIO_BY_ID.get(id);
  if (!scenario) {
    throw new Error(`Unknown scenario: ${id}`);
  }
  return scenario;
};

export const scenarioColor = (id: ScenarioId): string => {
  const severity = getScenarioMeta(id).severity;
  if (severity >= 0.75) return "#ff2b2b";
  if (severity >= 0.55) return "#ff8a00";
  if (severity >= 0.35) return "#ffd447";
  return "#39ff14";
};
