import { CRTOverlay } from "../components/CRTOverlay/CRTOverlay";
import { DoomsdayClock } from "../components/DoomsdayClock/DoomsdayClock";
import { GlobeScene } from "../components/GlobeScene/GlobeScene";
import { ScenarioPanel } from "../components/ScenarioPanel/ScenarioPanel";
import { SignalFeed } from "../components/SignalFeed/SignalFeed";
import { useClockData } from "../hooks/useClockData";
import { SCENARIOS, getScenarioMeta } from "../lib/scenarios";

const formatComputedAt = (computedAt: string) =>
  new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(new Date(computedAt));

export const Dashboard = () => {
  const { clock, signals, mode, lastError } = useClockData();
  const dominant = getScenarioMeta(clock.dominantScenario);
  const dominantWeight = clock.scenarioWeights[dominantIndex(clock.dominantScenario)] ?? 0;

  return (
    <main className="app-shell">
      <CRTOverlay />
      <SignalFeed signals={signals} />
      <section className="hero-grid">
        <div className="mission-panel">
          <p className="eyebrow">Life 3.0 doomsday clock</p>
          <h1>AI scenario risk monitor</h1>
          <p className="mission-copy">
            Tracking the system pressure across Max Tegmark's twelve long-run AI futures.
          </p>
          <div className="status-strip">
            <span className={`status-dot status-dot--${mode}`} />
            <span>{mode === "live" ? "Live GraphQL" : "Mock fallback"}</span>
            <span>{formatComputedAt(clock.computedAt)}</span>
          </div>
          {lastError ? <p className="fallback-note">{lastError}</p> : null}
        </div>
        <GlobeScene minutesToMidnight={clock.minutesToMidnight} />
        <DoomsdayClock clock={clock} />
        <aside className="dominant-panel">
          <p className="eyebrow">Dominant trajectory</p>
          <h2>{dominant.label}</h2>
          <p>{dominant.summary}</p>
          <dl>
            <div>
              <dt>Severity</dt>
              <dd>{Math.round(dominant.severity * 100)}%</dd>
            </div>
            <div>
              <dt>Weight</dt>
              <dd>{Math.round(dominantWeight * 100)}%</dd>
            </div>
          </dl>
        </aside>
      </section>
      <ScenarioPanel clock={clock} />
    </main>
  );
};

const dominantIndex = (scenarioId: string) => {
  return SCENARIOS.findIndex((scenario) => scenario.id === scenarioId);
};
