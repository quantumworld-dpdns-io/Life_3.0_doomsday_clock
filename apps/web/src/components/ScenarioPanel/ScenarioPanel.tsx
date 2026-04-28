import { SCENARIOS, getScenarioMeta, scenarioColor } from "../../lib/scenarios";
import type { ClockState } from "../../lib/types";

interface ScenarioPanelProps {
  clock: ClockState;
}

export const ScenarioPanel = ({ clock }: ScenarioPanelProps) => (
  <section className="scenario-panel" aria-label="Scenario probability panel">
    <div className="panel-heading">
      <div>
        <p className="eyebrow">Tegmark scenario distribution</p>
        <h2>{getScenarioMeta(clock.dominantScenario).label}</h2>
      </div>
      <span className="dominant-badge">Dominant</span>
    </div>
    <div className="scenario-table" role="table" aria-label="Scenario weights">
      <div className="scenario-row scenario-row--header" role="row">
        <span role="columnheader">Scenario</span>
        <span role="columnheader">Weight</span>
        <span role="columnheader">Confidence</span>
      </div>
      {SCENARIOS.map((scenario, index) => {
        const weight = Math.max(0, Math.min(1, clock.scenarioWeights[index] ?? 0));
        const isDominant = scenario.id === clock.dominantScenario;
        return (
          <div
            key={scenario.id}
            className={`scenario-row ${isDominant ? "scenario-row--dominant" : ""}`}
            role="row"
          >
            <div className="scenario-name" role="cell">
              <span className="scenario-index">{String(index + 1).padStart(2, "0")}</span>
              <span>
                <strong>{scenario.label}</strong>
                <small>{scenario.summary}</small>
              </span>
            </div>
            <div className="scenario-weight" role="cell">
              <span style={{ width: `${weight * 100}%`, backgroundColor: scenarioColor(scenario.id) }} />
            </div>
            <span className="scenario-confidence" role="cell">
              {Math.round(weight * 100)}%
            </span>
          </div>
        );
      })}
    </div>
  </section>
);
