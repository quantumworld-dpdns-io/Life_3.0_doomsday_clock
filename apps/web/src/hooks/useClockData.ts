import { useEffect, useMemo, useState } from "react";
import {
  fallbackClockState,
  fallbackSignals,
  fetchClockState,
  fetchRecentSignals,
  subscribeClockState,
} from "../lib/graphqlClient";
import { makeMockClockState, makeMockSignals } from "../lib/mockData";
import type { ClockState, DataMode, ScenarioSignal } from "../lib/types";

interface ClockData {
  clock: ClockState;
  signals: ScenarioSignal[];
  mode: DataMode;
  lastError?: string;
}

export const useClockData = (): ClockData => {
  const [clock, setClock] = useState<ClockState>(() => fallbackClockState());
  const [signals, setSignals] = useState<ScenarioSignal[]>(() => fallbackSignals(20));
  const [mode, setMode] = useState<DataMode>("mock");
  const [lastError, setLastError] = useState<string>();

  useEffect(() => {
    let mounted = true;
    let unsubscribe: (() => void) | undefined;

    const loadLiveData = async () => {
      try {
        const [nextClock, nextSignals] = await Promise.all([
          fetchClockState(),
          fetchRecentSignals(20),
        ]);
        if (!mounted) return;
        setClock(nextClock);
        setSignals(nextSignals);
        setMode("live");
        setLastError(undefined);

        unsubscribe = subscribeClockState(
          (state) => {
            setClock(state);
            setMode("live");
          },
          (error) => {
            setLastError(error.message);
          },
        );
      } catch (error) {
        if (!mounted) return;
        setMode("mock");
        setLastError(error instanceof Error ? error.message : "Backend unavailable");
      }
    };

    void loadLiveData();

    return () => {
      mounted = false;
      unsubscribe?.();
    };
  }, []);

  useEffect(() => {
    if (mode !== "mock") return undefined;

    const interval = window.setInterval(() => {
      const now = new Date();
      setClock(makeMockClockState(now));
      setSignals(makeMockSignals(now, 20));
    }, 3_500);

    return () => window.clearInterval(interval);
  }, [mode]);

  return useMemo(
    () => ({
      clock,
      signals,
      mode,
      lastError,
    }),
    [clock, lastError, mode, signals],
  );
};
