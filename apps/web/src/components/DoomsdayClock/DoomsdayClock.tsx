import { useEffect, useMemo, useRef } from "react";
import type { ClockState } from "../../lib/types";

interface DoomsdayClockProps {
  clock: ClockState;
}

const polar = (angle: number, radius: number) => {
  const radians = (angle - 90) * (Math.PI / 180);
  return {
    x: 100 + Math.cos(radians) * radius,
    y: 100 + Math.sin(radians) * radius,
  };
};

export const DoomsdayClock = ({ clock }: DoomsdayClockProps) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const minuteAngle = useMemo(() => (60 - clock.minutesToMidnight) * 6, [clock.minutesToMidnight]);
  const hourAngle = useMemo(() => 330 + (60 - clock.minutesToMidnight) * 0.5, [clock.minutesToMidnight]);
  const secondAngle = useMemo(() => {
    const jitter = Math.sin(Date.parse(clock.computedAt) / 137) * 5;
    return minuteAngle + jitter;
  }, [clock.computedAt, minuteAngle]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const context = canvas.getContext("2d");
    if (!context) return;

    const imageData = context.createImageData(canvas.width, canvas.height);
    for (let index = 0; index < imageData.data.length; index += 4) {
      const value = 18 + Math.random() * 42;
      imageData.data[index] = value;
      imageData.data[index + 1] = value;
      imageData.data[index + 2] = value;
      imageData.data[index + 3] = 32;
    }
    context.putImageData(imageData, 0, 0);
  }, []);

  const minuteHand = polar(minuteAngle, 74);
  const hourHand = polar(hourAngle, 44);
  const secondHand = polar(secondAngle, 82);

  return (
    <section className="clock-panel" aria-label="Doomsday clock">
      <canvas ref={canvasRef} className="clock-noise" width="280" height="280" aria-hidden="true" />
      <svg className="clock-face" viewBox="0 0 200 200" role="img" aria-label={`${clock.minutesToMidnight.toFixed(1)} minutes to midnight`}>
        <defs>
          <filter id="redGlow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="2.5" result="coloredBlur" />
            <feMerge>
              <feMergeNode in="coloredBlur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>
        <circle className="clock-ring" cx="100" cy="100" r="88" />
        <circle className="clock-inner" cx="100" cy="100" r="74" />
        {Array.from({ length: 60 }, (_, tick) => {
          const outer = polar(tick * 6, tick % 5 === 0 ? 86 : 84);
          const inner = polar(tick * 6, tick % 5 === 0 ? 75 : 80);
          return (
            <line
              key={tick}
              className={tick % 5 === 0 ? "clock-tick clock-tick--major" : "clock-tick"}
              x1={inner.x}
              y1={inner.y}
              x2={outer.x}
              y2={outer.y}
            />
          );
        })}
        <text className="clock-midnight-label" x="100" y="42" textAnchor="middle">
          MIDNIGHT
        </text>
        <line className="clock-hand clock-hand--hour" x1="100" y1="100" x2={hourHand.x} y2={hourHand.y} />
        <line className="clock-hand clock-hand--minute" x1="100" y1="100" x2={minuteHand.x} y2={minuteHand.y} />
        <line className="clock-hand clock-hand--second" x1="100" y1="100" x2={secondHand.x} y2={secondHand.y} />
        <circle className="clock-pin" cx="100" cy="100" r="4.5" />
      </svg>
      <div className="clock-readout">
        <span>{clock.minutesToMidnight.toFixed(1)}</span>
        <small>minutes to midnight</small>
      </div>
    </section>
  );
};
