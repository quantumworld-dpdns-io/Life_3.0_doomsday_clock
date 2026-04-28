import { motion } from "framer-motion";
import { getScenarioMeta, scenarioColor } from "../../lib/scenarios";
import type { ScenarioSignal } from "../../lib/types";

interface SignalFeedProps {
  signals: ScenarioSignal[];
}

const formatTime = (timestamp: string) =>
  new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(timestamp));

export const SignalFeed = ({ signals }: SignalFeedProps) => {
  const feed = signals.length > 0 ? signals : [];
  const items = [...feed, ...feed];

  return (
    <section className="signal-feed" aria-label="Recent scenario signals">
      <div className="signal-feed__label">Signal feed</div>
      <div className="signal-feed__track">
        <motion.div
          className="signal-feed__items"
          animate={{ x: ["0%", "-50%"] }}
          transition={{ repeat: Infinity, duration: 42, ease: "linear" }}
        >
          {items.map((signal, index) => {
            const scenario = getScenarioMeta(signal.scenario);
            return (
              <a
                key={`${signal.sourceUrl}-${index}`}
                className="signal-item"
                href={signal.sourceUrl.startsWith("mock://") ? undefined : signal.sourceUrl}
                style={{ borderColor: scenarioColor(signal.scenario) }}
                target="_blank"
                rel="noreferrer"
              >
                <span>{formatTime(signal.timestamp)}</span>
                <strong>{scenario.label}</strong>
                <em>{Math.round(signal.confidence * 100)}%</em>
                <small>{signal.title ?? signal.sourceUrl}</small>
              </a>
            );
          })}
        </motion.div>
      </div>
    </section>
  );
};
