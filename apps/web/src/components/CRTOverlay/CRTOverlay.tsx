import { useEffect, useState } from "react";

export const CRTOverlay = () => {
  const [glitching, setGlitching] = useState(false);

  useEffect(() => {
    let timeoutId: number;

    const schedule = () => {
      const delay = 8_000 + Math.random() * 22_000;
      timeoutId = window.setTimeout(() => {
        setGlitching(true);
        window.setTimeout(() => setGlitching(false), 80);
        schedule();
      }, delay);
    };

    schedule();
    return () => window.clearTimeout(timeoutId);
  }, []);

  return <div className={`crt-overlay ${glitching ? "crt-overlay--glitch" : ""}`} />;
};
