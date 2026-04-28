import type { ScenarioId } from "./scenarios";
import { makeMockClockState, makeMockSignals } from "./mockData";
import type { ClockState, ScenarioSignal } from "./types";

const DEFAULT_HTTP_URL = "http://localhost:4000/graphql";
const DEFAULT_WS_URL = "ws://localhost:4000/graphql";

const CLOCK_QUERY = `
  query ClockState {
    clockState {
      minutesToMidnight
      dominantScenario
      scenarioWeights
      computedAt
    }
  }
`;

const SIGNALS_QUERY = `
  query RecentSignals($limit: Int = 20) {
    recentSignals(limit: $limit) {
      scenario
      confidence
      sourceUrl
      timestamp
    }
  }
`;

const CLOCK_SUBSCRIPTION = `
  subscription ClockStateStream {
    clockStateStream {
      minutesToMidnight
      dominantScenario
      scenarioWeights
      computedAt
    }
  }
`;

interface GraphQLResponse<T> {
  data?: T;
  errors?: { message: string }[];
}

const env = import.meta.env;

export const graphqlHttpUrl = env.VITE_GRAPHQL_HTTP || DEFAULT_HTTP_URL;
export const graphqlWsUrl = env.VITE_GRAPHQL_WS || DEFAULT_WS_URL;

const request = async <T>(query: string, variables?: Record<string, unknown>): Promise<T> => {
  const response = await fetch(graphqlHttpUrl, {
    method: "POST",
    headers: {
      "content-type": "application/json",
      ...(env.VITE_GRAPHQL_TOKEN ? { authorization: `Bearer ${env.VITE_GRAPHQL_TOKEN}` } : {}),
    },
    body: JSON.stringify({ query, variables }),
  });

  if (!response.ok) {
    throw new Error(`GraphQL HTTP ${response.status}`);
  }

  const payload = (await response.json()) as GraphQLResponse<T>;
  if (payload.errors?.length) {
    throw new Error(payload.errors.map((error) => error.message).join("; "));
  }
  if (!payload.data) {
    throw new Error("GraphQL response did not include data");
  }
  return payload.data;
};

const normalizeClock = (state: ClockState): ClockState => ({
  minutesToMidnight: Number(state.minutesToMidnight),
  dominantScenario: state.dominantScenario,
  scenarioWeights: state.scenarioWeights.map(Number),
  computedAt: state.computedAt,
});

const normalizeSignal = (signal: ScenarioSignal): ScenarioSignal => ({
  scenario: signal.scenario,
  confidence: Number(signal.confidence),
  sourceUrl: signal.sourceUrl,
  timestamp: signal.timestamp,
  title: signal.title,
});

export const fetchClockState = async (): Promise<ClockState> => {
  const data = await request<{ clockState: ClockState }>(CLOCK_QUERY);
  return normalizeClock(data.clockState);
};

export const fetchRecentSignals = async (limit = 20): Promise<ScenarioSignal[]> => {
  const data = await request<{ recentSignals: ScenarioSignal[] }>(SIGNALS_QUERY, { limit });
  return data.recentSignals.map(normalizeSignal);
};

export const fallbackClockState = (): ClockState => makeMockClockState();

export const fallbackSignals = (limit = 20): ScenarioSignal[] => makeMockSignals(new Date(), limit);

export const subscribeClockState = (
  onState: (state: ClockState) => void,
  onError: (error: Error) => void,
): (() => void) => {
  let socket: WebSocket | undefined;
  let closed = false;

  try {
    socket = new WebSocket(graphqlWsUrl, "graphql-transport-ws");
  } catch (error) {
    onError(error instanceof Error ? error : new Error("Unable to open GraphQL WebSocket"));
    return () => {
      closed = true;
    };
  }

  socket.addEventListener("open", () => {
    socket?.send(JSON.stringify({ type: "connection_init" }));
    socket?.send(
      JSON.stringify({
        id: "clock-state-stream",
        type: "subscribe",
        payload: { query: CLOCK_SUBSCRIPTION },
      }),
    );
  });

  socket.addEventListener("message", (event) => {
    const payload = JSON.parse(String(event.data)) as {
      type: string;
      payload?: { data?: { clockStateStream?: ClockState } };
    };
    const state = payload.payload?.data?.clockStateStream;
    if (state) {
      onState(normalizeClock(state));
    }
  });

  socket.addEventListener("error", () => {
    if (!closed) {
      onError(new Error("GraphQL WebSocket unavailable"));
    }
  });

  return () => {
    closed = true;
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ id: "clock-state-stream", type: "complete" }));
    }
    socket?.close();
  };
};

export const isScenarioId = (value: string): value is ScenarioId =>
  [
    "LIBERTARIAN_UTOPIA",
    "BENEVOLENT_DICTATOR",
    "EGALITARIAN_UTOPIA",
    "GATEKEEPER",
    "PROTECTOR_GOD",
    "ENSLAVED_GOD",
    "CONQUERORS",
    "DESCENDANTS",
    "ZOOKEEPER",
    "NINETEEN_EIGHTY_FOUR",
    "REVERT",
    "SELF_DESTRUCTION",
  ].includes(value);
