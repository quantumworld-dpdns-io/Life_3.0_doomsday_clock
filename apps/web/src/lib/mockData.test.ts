import { describe, expect, it } from "vitest";
import { SCENARIOS } from "./scenarios";
import { makeMockClockState, makeMockSignals } from "./mockData";

describe("mock data fallback", () => {
  it("creates a complete clock state", () => {
    const clock = makeMockClockState(new Date("2026-04-28T12:00:00.000Z"));

    expect(clock.minutesToMidnight).toBeGreaterThan(0);
    expect(clock.minutesToMidnight).toBeLessThanOrEqual(60);
    expect(clock.scenarioWeights).toHaveLength(SCENARIOS.length);
    expect(SCENARIOS.map((scenario) => scenario.id)).toContain(clock.dominantScenario);
  });

  it("creates recent signals with scenario confidence", () => {
    const signals = makeMockSignals(new Date("2026-04-28T12:00:00.000Z"), 6);

    expect(signals).toHaveLength(6);
    expect(signals[0].confidence).toBeGreaterThan(0);
    expect(signals[0].sourceUrl).toContain("mock://life3/signals");
  });
});
