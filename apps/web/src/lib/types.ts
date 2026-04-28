import type { ScenarioId } from "./scenarios";

export interface ClockState {
  minutesToMidnight: number;
  dominantScenario: ScenarioId;
  scenarioWeights: number[];
  computedAt: string;
}

export interface ScenarioSignal {
  scenario: ScenarioId;
  confidence: number;
  sourceUrl: string;
  timestamp: string;
  title?: string;
}

export type DataMode = "live" | "mock";
