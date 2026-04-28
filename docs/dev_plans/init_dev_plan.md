# Life 3.0 Doomsday Clock — Detailed Development Plan

> Based on `init_spec.md`. Covers architecture, directory scaffold, service-by-service implementation, integration, and frontend.

---

## 0. Project Overview

A polyglot monorepo that tracks the world's proximity to Max Tegmark's 12 AI-evolution scenarios (from *Life 3.0*, 2017). The system ingests global AI/news signals, classifies them against Tegmark's taxonomy, runs a probabilistic "doomsday clock" simulation (with genuine quantum randomness), and surfaces results through a dark, heavy-aesthetic React SPA.

**Data flow:**
```
Reuters / arXiv / HN
        │
        ▼
[Python] Scraper + LLM Classifier
        │  (scenario label + confidence score)
        ▼
[Rust]  Probability Engine  ←──  [QASM] Quantum Entropy Seed
        │  (weighted clock state)
        ▼
[Go]    API Gateway (gRPC / GraphQL)
        │
        ▼
[React] Dark UI  (Three.js + nothingx-react-components)
```

---

## Current Implementation Snapshot

**Observed:** 2026-04-28 from the current working tree.

This section tracks implementation progress without replacing the target architecture below. Items are marked present only when corresponding files or directories currently exist in the repo.

| Area | Current status | Evidence in tree | Pending verification / gaps |
|---|---|---|---|
| Repository scaffold | Partially present | `services/`, `apps/`, `shared/`, `docs/`, `.github/workflows/ci.yaml`, `Makefile`, `docker-compose.yaml`, `docker-compose.override.yaml` | Some service/app directories exist without source files; full cold-start has not been verified in this update. |
| Shared proto contracts | Present | `shared/proto/clock.proto`, `shared/proto/scenario.proto`, `shared/proto/entropy.proto`, `shared/scripts/gen_proto.sh` | Regenerated stubs and `git diff --exit-code` still need to be run in CI/local verification. |
| Intelligence service | Implemented scaffold with app modules and tests present | `services/intelligence/pyproject.toml`, `Dockerfile`, `app/api.py`, `app/grpc_server.py`, scraper modules, classifier modules, DB migration, tests | Need run `uv run pytest`/lint and verify scraper/classifier behavior against real or mocked sources. |
| Quantum entropy service | Implemented scaffold with app modules and tests present | `services/quantum-sim/pyproject.toml`, `Dockerfile`, `app/runner.py`, `app/api.py`, `app/grpc_server.py`, `app/circuits/entropy.qasm`, tests | Need run quantum tests and confirm native Qiskit/Aer dependencies build in Docker. |
| Risk engine | Implemented scaffold with Rust source present | `services/risk-engine/Cargo.toml`, `build.rs`, `Dockerfile`, `src/main.rs`, `src/grpc_server.rs`, `src/clock/`, `src/monte_carlo/`, `src/clients.rs`, `config/scenario_weights.toml` | Need run `cargo test`/`cargo clippy`; verify generated proto bindings and live calls to intelligence/quantum services. |
| API gateway | Directory scaffold only | `services/api-gateway/cmd/gateway/`, `internal/auth/`, `internal/graphql/`, `internal/grpc_client/` directories exist | No source files were present under `services/api-gateway` during this snapshot; Go module, Dockerfile, GraphQL schema/resolvers, and tests remain to be added or verified if another worker is still writing them. |
| Frontend web app | Early Vite/source scaffold present | `apps/web/package.json`, `Dockerfile`, `index.html`, Vite/Tailwind/TypeScript config files, `src/vite-env.d.ts`, `src/lib/*`, `src/hooks/useClockData.ts`, and component/source directories exist | No component implementation files were present under `apps/web/src/components` during this snapshot; rendered UI, tests, and runnable frontend behavior remain to be verified if another worker is still writing them. |
| Compose/dev commands | Planned wiring present | `docker-compose.yaml` references Postgres, NATS, intelligence, quantum-sim, risk-engine, api-gateway, and web; `Makefile` has `proto`, `build`, `up`, `down`, `logs`, `lint`, `test`, `migrate`, `clean`, `help` | `docker compose build`/`up`, `make lint`, and `make test` are pending verification; current compose and Make targets reference gateway/web paths that do not yet contain implementation files. |
| CI | Workflow present | `.github/workflows/ci.yaml` contains proto, lint, and test jobs for Go, Python, Rust, and frontend | CI likely needs updates once gateway/web source files are committed; current jobs may fail on empty gateway/web directories. |

### Phase Progress

| Phase | Progress | Notes |
|---|---|---|
| 0 - Scaffold | Partial | Core monorepo directories, compose, Makefile, CI, shared protos, and service directories are present. Gateway/web are directory-only at this snapshot. |
| 1 - Data Pipeline | Partial | Intelligence service files for API, DB, scrapers, classifier, scheduler, gRPC, migration, Dockerfile, and tests are present. Runtime behavior has not been verified here. |
| 2 - Entropy | Partial | Quantum service files, QASM circuit, runner/API/gRPC modules, Dockerfile, and tests are present. Entropy distribution and Docker build remain pending verification. |
| 3 - Risk Engine | Partial | Rust engine source, config, build script, Dockerfile, and gRPC/clock/Monte Carlo module paths are present. Tests and inter-service calls remain pending verification. |
| 4 - Gateway | Not yet implemented in files | Directory skeleton exists, but no Go source/module files were present during this snapshot. |
| 5 - Frontend MVP | Early scaffold | Vite/package/config files plus frontend lib/hook files and component directories exist, but no component implementation files were present under `apps/web/src/components` during this snapshot. |
| 6 - Polish | Not started | CRT/glitch/shader/ticker features remain planned until frontend source exists. |
| 7 - Hardening | Not started | Production hardening should follow a verified runnable stack. |

---

## 1. Repository Scaffold

### 1.1 Target Directory Tree

```
life-3.0-doomsday-clock/
├── apps/
│   ├── web/                          # React SPA
│   │   ├── src/
│   │   │   ├── components/
│   │   │   │   ├── DoomsdayClock/
│   │   │   │   ├── ScenarioPanel/
│   │   │   │   ├── GlobeScene/       # Three.js low-poly sphere
│   │   │   │   └── CRTOverlay/
│   │   │   ├── hooks/
│   │   │   ├── pages/
│   │   │   ├── store/                # Zustand or Redux
│   │   │   └── main.tsx
│   │   ├── public/
│   │   ├── index.html
│   │   ├── tailwind.config.ts
│   │   ├── vite.config.ts
│   │   └── package.json
│   └── docs/                         # Tegmark theory notes, ADRs
├── services/
│   ├── api-gateway/                  # [Go]
│   │   ├── cmd/gateway/main.go
│   │   ├── internal/
│   │   │   ├── auth/
│   │   │   ├── graphql/
│   │   │   └── grpc_client/
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── intelligence/                 # [Python]
│   │   ├── app/
│   │   │   ├── scraper/
│   │   │   ├── classifier/
│   │   │   └── api.py               # FastAPI entry
│   │   ├── pyproject.toml
│   │   └── Dockerfile
│   ├── risk-engine/                  # [Rust]
│   │   ├── src/
│   │   │   ├── main.rs
│   │   │   ├── clock/
│   │   │   ├── monte_carlo/
│   │   │   └── grpc/
│   │   ├── proto/                    # local copy (symlink from shared/)
│   │   ├── Cargo.toml
│   │   └── Dockerfile
│   └── quantum-sim/                  # [Python + QASM]
│       ├── app/
│       │   ├── circuits/
│       │   │   └── entropy.qasm
│       │   ├── runner.py
│       │   └── api.py               # FastAPI entry
│       ├── pyproject.toml
│       └── Dockerfile
├── shared/
│   ├── proto/
│   │   ├── clock.proto
│   │   ├── scenario.proto
│   │   └── entropy.proto
│   └── scripts/
│       ├── gen_proto.sh
│       ├── seed_db.sh
│       └── ci_check.sh
├── docker-compose.yaml
├── docker-compose.override.yaml      # local dev overrides
├── Makefile
├── .github/
│   └── workflows/
│       ├── ci.yaml
│       └── release.yaml
└── CLAUDE.md
```

### 1.2 Scaffold Tasks

| # | Task | Owner |
|---|------|-------|
| S1 | `mkdir -p` all directories above | DevOps |
| S2 | Create root `Makefile` with targets: `build`, `up`, `down`, `proto`, `lint`, `test` | DevOps |
| S3 | Create root `docker-compose.yaml` wiring all 4 services + Postgres + NATS | DevOps |
| S4 | Add `.github/workflows/ci.yaml` running lint + test per service on PR | DevOps |
| S5 | Initialize each service's package manager (`go mod init`, `cargo init`, `uv init`) | Per-service |

---

## 2. Shared Proto Definitions

All inter-service communication uses gRPC. Define the contracts first so every team can develop in parallel.

### 2.1 `shared/proto/scenario.proto`

```protobuf
syntax = "proto3";
package life3;

enum ScenarioID {
  UNKNOWN            = 0;
  LIBERTARIAN_UTOPIA = 1;
  BENEVOLENT_DICTATOR= 2;
  EGALITARIAN_UTOPIA = 3;
  GATEKEEPER         = 4;
  PROTECTOR_GOD      = 5;
  ENSLAVED_GOD       = 6;
  CONQUERORS         = 7;
  DESCENDANTS        = 8;
  ZOOKEEPER          = 9;
  1984               = 10;
  REVERT             = 11;
  SELF_DESTRUCTION   = 12;
}

message ScenarioSignal {
  ScenarioID   scenario    = 1;
  float        confidence  = 2;   // 0.0–1.0
  string       source_url  = 3;
  int64        timestamp   = 4;
}
```

### 2.2 `shared/proto/entropy.proto`

```protobuf
syntax = "proto3";
package life3;

message EntropyRequest {
  int32 num_bits = 1;
}

message EntropyResponse {
  repeated bool bits   = 1;
  float        delta   = 2;   // pre-computed ±variance contribution
}
```

### 2.3 `shared/proto/clock.proto`

```protobuf
syntax = "proto3";
package life3;

message ClockState {
  float          minutes_to_midnight = 1;   // e.g. 1.5
  ScenarioID     dominant_scenario   = 2;
  float          scenario_confidence = 3;
  repeated float scenario_weights    = 4;   // length 12, index = ScenarioID
  int64          computed_at         = 5;
}

service RiskEngine {
  rpc GetClockState  (google.protobuf.Empty) returns (ClockState);
  rpc StreamClockState (google.protobuf.Empty) returns (stream ClockState);
}
```

### 2.4 Code-gen

`shared/scripts/gen_proto.sh` runs `protoc` targeting Go, Python, and Rust (via `tonic-build`). Add it to `make proto`.

---

## 3. Service A — Intelligence Layer (Python)

**Location:** `services/intelligence/`
**Stack:** Python 3.12, FastAPI, LangChain, APScheduler, httpx, PostgreSQL (asyncpg)

### 3.1 Scraper Sub-module

**File:** `app/scraper/sources.py`

| Source | Method | Cadence |
|--------|--------|---------|
| Reuters AI tag | RSS feed | every 15 min |
| arXiv cs.AI | API (`arxiv` lib) | every 1 h |
| Hacker News | Algolia Search API | every 30 min |
| Tegmark's FLI news | RSS | every 1 h |

Each scraper returns a normalized `Article(url, title, body, published_at)` dataclass. Store raw articles in Postgres table `raw_articles`.

### 3.2 LLM Classifier Sub-module

**File:** `app/classifier/llm.py`

- Use `langchain_openai.ChatOpenAI` (or `langchain_community.llms.Ollama` for local Llama-3).
- Prompt strategy: few-shot with one canonical example per Tegmark scenario (store examples in `app/classifier/prompts/few_shot.yaml`).
- Output: structured JSON via `langchain_core.output_parsers.JsonOutputParser` mapping to `ScenarioSignal`.
- Batch articles every scrape cycle; skip if already classified (idempotency by URL hash).

**Classifier prompt skeleton:**

```
You are an AI risk analyst. Given a news article, classify it into exactly one
of the 12 Life 3.0 scenarios defined by Max Tegmark. Return JSON:
{"scenario": "<NAME>", "confidence": <0.0-1.0>, "reasoning": "<one sentence>"}

Scenarios:
1. Libertarian Utopia – ...
...
12. Self-Destruction – ...

Article:
{article_text}
```

### 3.3 FastAPI Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/signals/latest` | Returns the 50 most recent `ScenarioSignal` rows |
| GET | `/signals/aggregate` | Returns weighted scenario distribution (last 7 days) |
| POST | `/scrape/trigger` | Admin-only manual trigger |
| GET | `/health` | Liveness probe |

### 3.4 gRPC Server

Expose `ScenarioSignalService` (defined in `scenario.proto`) so the Rust risk engine can subscribe via streaming RPC instead of polling REST.

### 3.5 Database Schema

```sql
CREATE TABLE raw_articles (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url          TEXT UNIQUE NOT NULL,
  title        TEXT,
  body         TEXT,
  published_at TIMESTAMPTZ,
  ingested_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE scenario_signals (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id   UUID REFERENCES raw_articles(id),
  scenario     SMALLINT NOT NULL,   -- ScenarioID enum value
  confidence   REAL NOT NULL,
  created_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX ON scenario_signals (created_at DESC);
```

### 3.6 Implementation Steps

| # | Step |
|---|------|
| I1 | `uv init` + add deps: `fastapi uvicorn langchain langchain-openai arxiv httpx asyncpg apscheduler` |
| I2 | Implement `Article` dataclass and Postgres repository layer |
| I3 | Implement Reuters, arXiv, HN scrapers; write unit tests with `pytest` + `respx` |
| I4 | Implement few-shot prompt YAML and `LLMClassifier` class |
| I5 | Wire APScheduler jobs in `app/scheduler.py` |
| I6 | Implement FastAPI app in `app/api.py` |
| I7 | Implement gRPC server in `app/grpc_server.py` using `grpcio` |
| I8 | Write Dockerfile: multi-stage, `python:3.12-slim` base |
| I9 | Add integration test hitting a local Postgres fixture |

---

## 4. Service B — Quantum Entropy Simulator (Python + QASM)

**Location:** `services/quantum-sim/`
**Stack:** Python 3.12, FastAPI, Qiskit, `qiskit-aer` (local simulator)

### 4.1 QASM Circuit

**File:** `app/circuits/entropy.qasm`

```qasm
OPENQASM 2.0;
include "qelib1.inc";

qreg q[4];
creg c[4];

// Hadamard on all qubits → superposition → collapse = true randomness
h q[0];
h q[1];
h q[2];
h q[3];

measure q[0] -> c[0];
measure q[1] -> c[1];
measure q[2] -> c[2];
measure q[3] -> c[3];
```

Each shot produces 4 random bits. Run N shots to fill an entropy request.

### 4.2 Runner

**File:** `app/runner.py`

```python
from qiskit import QuantumCircuit, transpile
from qiskit_aer import AerSimulator

def get_entropy_bits(num_bits: int) -> list[bool]:
    shots = (num_bits + 3) // 4          # 4 bits per shot
    qc    = QuantumCircuit.from_qasm_file("circuits/entropy.qasm")
    sim   = AerSimulator()
    job   = sim.run(transpile(qc, sim), shots=shots)
    counts = job.result().get_counts()
    bits: list[bool] = []
    for bitstring, freq in counts.items():
        for _ in range(freq):
            bits.extend(b == "1" for b in bitstring)
    return bits[:num_bits]
```

### 4.3 gRPC Server

Implement `EntropyService` (from `entropy.proto`). The Rust engine calls this when it needs entropy for a Monte Carlo step.

### 4.4 Implementation Steps

| # | Step |
|---|------|
| Q1 | `uv init` + add deps: `qiskit qiskit-aer fastapi uvicorn grpcio protobuf` |
| Q2 | Write and test `entropy.qasm` circuit using Qiskit's local `AerSimulator` |
| Q3 | Implement `runner.py` with unit tests verifying bit count and type |
| Q4 | Implement gRPC `EntropyService` server |
| Q5 | Dockerfile: `python:3.12-slim` + install `qiskit-aer` native deps |
| Q6 | Verify entropy distribution is approximately uniform (chi-squared test in CI) |

---

## 5. Service C — Risk Engine (Rust)

**Location:** `services/risk-engine/`
**Stack:** Rust 1.78+, Axum, Tonic (gRPC), Tokio, `rand`, `nalgebra`

### 5.1 Clock State Model

```rust
// src/clock/state.rs
pub struct ClockState {
    pub minutes_to_midnight: f32,      // 0.0 (midnight) – 60.0 (safe)
    pub scenario_weights: [f32; 13],   // index 0 unused; 1–12 = ScenarioID
    pub dominant_scenario: u8,
    pub computed_at: i64,
}
```

### 5.2 Weight Computation

Each Tegmark scenario has a **base severity score** (hand-tuned from the book, stored in `config/scenario_weights.toml`):

```toml
[scenarios]
libertarian_utopia  = 0.10
benevolent_dictator = 0.20
gatekeeper          = 0.40
conquerors          = 0.70
self_destruction    = 1.00
# ...
```

The engine fetches the latest `scenario_weights` aggregate from the intelligence service and multiplies the LLM confidence scores by the base severity:

```
effective_weight[i] = llm_confidence[i] * base_severity[i]
minutes_to_midnight = 60.0 * (1.0 - Σ effective_weight[i])
```

### 5.3 Monte Carlo Simulation

**File:** `src/monte_carlo/sim.rs`

- Run 10,000 iterations per clock tick.
- Each iteration: fetch N entropy bits from the quantum-sim service; use them to perturb `effective_weight[i]` by ±`delta`.
- Report `mean ± σ` of `minutes_to_midnight` across iterations.
- This gives the clock a probabilistic range rather than a point estimate.

### 5.4 gRPC Server (Tonic)

Implement `RiskEngine` service from `clock.proto`:
- `GetClockState`: single response with latest computed state.
- `StreamClockState`: server-streaming; pushes a new `ClockState` every 60 seconds (or on significant signal change).

### 5.5 Implementation Steps

| # | Step |
|---|------|
| R1 | `cargo init` + add deps: `tonic axum tokio serde serde_json config rand nalgebra` |
| R2 | Run `tonic-build` in `build.rs` to generate gRPC stubs from `clock.proto` + `entropy.proto` |
| R3 | Implement `ScenarioWeightConfig` loader from `config/scenario_weights.toml` |
| R4 | Implement gRPC client for intelligence service (fetches latest aggregate) |
| R5 | Implement gRPC client for quantum-sim service (fetches entropy bits) |
| R6 | Implement `monte_carlo::simulate()` using entropy bits as RNG seed |
| R7 | Implement `ClockState` computation combining steps R3–R6 |
| R8 | Implement Tonic gRPC server for `RiskEngine` |
| R9 | Add Axum HTTP health endpoint at `GET /health` |
| R10 | Unit tests for weight math; property tests for Monte Carlo distribution |
| R11 | Dockerfile: `rust:1.78-slim` builder + `debian:bookworm-slim` runtime |

---

## 6. Service D — API Gateway (Go)

**Location:** `services/api-gateway/`
**Stack:** Go 1.22+, `gqlgen` (GraphQL), `google.golang.org/grpc`, `golang-jwt/jwt`, NATS (optional pub/sub)

### 6.1 Responsibilities

- Single public ingress for the React SPA.
- Terminates JWT auth (issue tokens via `/auth/login`).
- Translates GraphQL queries/subscriptions → gRPC calls to risk-engine & intelligence.
- Optionally re-publishes `ClockState` updates to NATS for future consumers.

### 6.2 GraphQL Schema

```graphql
type Query {
  clockState: ClockState!
  recentSignals(limit: Int = 20): [ScenarioSignal!]!
}

type Subscription {
  clockStateStream: ClockState!
}

type ClockState {
  minutesToMidnight: Float!
  dominantScenario:  Scenario!
  scenarioWeights:   [Float!]!   # length 12
  computedAt:        String!
}

type ScenarioSignal {
  scenario:   Scenario!
  confidence: Float!
  sourceUrl:  String!
  timestamp:  String!
}

enum Scenario {
  LIBERTARIAN_UTOPIA
  BENEVOLENT_DICTATOR
  EGALITARIAN_UTOPIA
  GATEKEEPER
  PROTECTOR_GOD
  ENSLAVED_GOD
  CONQUERORS
  DESCENDANTS
  ZOOKEEPER
  NINETEEN_EIGHTY_FOUR
  REVERT
  SELF_DESTRUCTION
}
```

### 6.3 Auth Flow

1. `POST /auth/login` → validates API key from env → returns signed JWT (HS256, 24 h TTL).
2. All GraphQL requests require `Authorization: Bearer <token>`.
3. Middleware validates JWT before passing to resolver.

### 6.4 Implementation Steps

| # | Step |
|---|------|
| G1 | `go mod init` + add deps: `gqlgen grpc jwt nats.go` |
| G2 | Generate gRPC Go stubs from `shared/proto/` |
| G3 | Implement JWT middleware in `internal/auth/` |
| G4 | Run `gqlgen init`, define schema above, implement resolvers |
| G5 | Implement gRPC client pool for risk-engine (with retry/backoff) |
| G6 | Implement gRPC client for intelligence service |
| G7 | Implement GraphQL subscription using risk-engine's `StreamClockState` |
| G8 | Add NATS publisher (optional; behind feature flag) |
| G9 | Unit tests for auth middleware; integration tests with testcontainers |
| G10 | Dockerfile: multi-stage `golang:1.22-alpine` builder + `alpine` runtime |

---

## 7. Frontend — React SPA

**Location:** `apps/web/`
**Stack:** React 18, TypeScript, Vite, Tailwind CSS v4, Framer Motion, Three.js (via `@react-three/fiber` + `@react-three/drei`), `@dennislee928/nothingx-react-components`, Apollo Client (GraphQL + subscriptions via WebSocket)

### 7.1 Design System

| Token | Value |
|-------|-------|
| `--bg` | `#0D0D0D` |
| `--accent-red` | `#FF0000` |
| `--accent-green` | `#39FF14` |
| `--text-primary` | `#E0E0E0` |
| `--font-mono` | `"JetBrains Mono", "Courier New", monospace` |
| `--font-display` | `"Impact", "Inter Black", sans-serif` |

All colors defined as CSS custom properties in `src/styles/tokens.css`. Tailwind extended to use these tokens.

### 7.2 Key Components

#### `GlobeScene` (Three.js)

- Low-poly icosphere (`IcosahedronGeometry`, `detail=2`).
- Slow Y-axis rotation (0.001 rad/frame).
- Color interpolates from `#003300` (safe) → `#FF0000` (midnight) based on `minutesToMidnight`.
- At `minutesToMidnight < 2`, vertices begin random displacement (disintegration effect) using a vertex shader.
- Implemented in `src/components/GlobeScene/GlobeScene.tsx` using `@react-three/fiber`.

#### `DoomsdayClock`

- SVG analog clock face, hands driven by `minutesToMidnight`.
- Second hand jitters randomly (entropy feed from the quantum-sim gives micro-jitter values via Apollo subscription).
- CRT scanline overlay using a CSS `::before` pseudo-element with repeating linear gradient.
- Subtle noise texture on `<canvas>` background (Perlin noise generated once at mount).

#### `ScenarioPanel`

- 12-row table, one per Tegmark scenario.
- Each row shows: scenario name, weight bar (Tailwind `bg-accent-red` width %), confidence %.
- Rows from `@dennislee928/nothingx-react-components` for the data-table primitives.
- Dominant scenario row highlighted with a blinking border (`animate-pulse`).

#### `CRTOverlay`

- Full-screen fixed `<div>` with `pointer-events: none`.
- Scanline: CSS `repeating-linear-gradient(transparent, transparent 2px, rgba(0,0,0,0.15) 2px, rgba(0,0,0,0.15) 4px)`.
- Vignette: radial gradient dark edges.
- Occasional glitch: random CSS `translate` + `skew` applied for 80 ms every 8–30 s (random interval via `useEffect`).

#### `SignalFeed`

- Scrolling ticker of recent `ScenarioSignal` items.
- Implemented as a horizontally scrolling marquee using Framer Motion `animate={{ x: [0, -100%] }}` loop.
- Each item color-coded by scenario severity.

### 7.3 Apollo Client Setup

```typescript
// src/apollo.ts
import { ApolloClient, InMemoryCache, split, HttpLink } from "@apollo/client";
import { GraphQLWsLink } from "@apollo/client/link/subscriptions";
import { createClient } from "graphql-ws";
import { getMainDefinition } from "@apollo/client/utilities";

const wsLink   = new GraphQLWsLink(createClient({ url: "ws://localhost:4000/graphql" }));
const httpLink = new HttpLink({ uri: "http://localhost:4000/graphql" });

const splitLink = split(
  ({ query }) => {
    const def = getMainDefinition(query);
    return def.kind === "OperationDefinition" && def.operation === "subscription";
  },
  wsLink,
  httpLink
);

export const client = new ApolloClient({ link: splitLink, cache: new InMemoryCache() });
```

### 7.4 Implementation Steps

| # | Step |
|---|------|
| F1 | `npm create vite@latest web -- --template react-ts` |
| F2 | Install deps: `tailwindcss @react-three/fiber @react-three/drei three framer-motion @apollo/client graphql graphql-ws @dennislee928/nothingx-react-components` |
| F3 | Configure Tailwind with design tokens (extend colors, fontFamily) |
| F4 | Implement `CRTOverlay` component with scanlines + vignette + glitch effect |
| F5 | Implement `GlobeScene` with low-poly sphere, rotation, color interpolation |
| F6 | Implement `DoomsdayClock` SVG component connected to Apollo subscription |
| F7 | Implement `ScenarioPanel` using nothingx table primitives |
| F8 | Implement `SignalFeed` ticker using Framer Motion |
| F9 | Wire Apollo Client; connect all components to GraphQL queries/subscriptions |
| F10 | Add vertex shader for disintegration effect when clock near midnight |
| F11 | Add JWT auth flow (login page → token stored in memory, not localStorage) |
| F12 | Responsive layout for desktop (primary) and tablet |
| F13 | Lighthouse audit: target performance ≥ 70 (heavy 3D is expected) |

---

## 8. Infrastructure & DevOps

### 8.1 `docker-compose.yaml` Services

| Service | Image | Ports | Depends On |
|---------|-------|-------|------------|
| `postgres` | `postgres:16-alpine` | 5432 | — |
| `nats` | `nats:2.10-alpine` | 4222, 8222 | — |
| `intelligence` | `./services/intelligence` | 8001 (HTTP), 50051 (gRPC) | postgres |
| `quantum-sim` | `./services/quantum-sim` | 8002 (HTTP), 50052 (gRPC) | — |
| `risk-engine` | `./services/risk-engine` | 8003 (HTTP), 50053 (gRPC) | intelligence, quantum-sim |
| `api-gateway` | `./services/api-gateway` | 4000 | risk-engine, intelligence, nats |
| `web` | `./apps/web` | 3000 | api-gateway |

### 8.2 `Makefile` Targets

```makefile
proto:        ## Generate gRPC stubs for all languages
    bash shared/scripts/gen_proto.sh

build:        ## Build all Docker images
    docker compose build

up:           ## Start all services
    docker compose up -d

down:         ## Stop all services
    docker compose down

lint:         ## Run linters (golangci-lint, ruff, clippy, eslint)
    ...

test:         ## Run all unit tests
    ...

migrate:      ## Run Postgres migrations (via golang-migrate)
    ...
```

### 8.3 CI Pipeline (`.github/workflows/ci.yaml`)

On every PR:
1. Lint each service in parallel (golangci-lint / ruff / clippy / eslint).
2. Unit tests per service (with Postgres + NATS testcontainers for integration tests).
3. Proto consistency check (gen_proto.sh + `git diff --exit-code`).
4. Docker build smoke test.

---

## 9. Phase Roadmap

| Phase | Deliverables | Estimated Effort |
|-------|-------------|-----------------|
| **0 — Scaffold** | Directory tree, Makefile, docker-compose, CI skeleton, proto files | 1–2 days |
| **1 — Data Pipeline** | Intelligence scraper + LLM classifier running + signals in Postgres | 3–5 days |
| **2 — Entropy** | Quantum-sim QASM circuit + gRPC service returning entropy bits | 1–2 days |
| **3 — Risk Engine** | Rust Monte Carlo engine consuming signals + entropy, streaming ClockState | 3–4 days |
| **4 — Gateway** | Go API Gateway: auth + GraphQL + subscriptions wired to risk-engine | 2–3 days |
| **5 — Frontend MVP** | Globe, Clock, ScenarioPanel connected to live data | 4–6 days |
| **6 — Polish** | CRT effects, glitch, disintegration shader, signal ticker, responsive | 2–3 days |
| **7 — Hardening** | Rate limiting, secrets management, production Dockerfiles, load test | 2–3 days |

---

## 10. Key Decisions & Open Questions

| # | Decision / Question | Recommendation |
|---|---------------------|---------------|
| D1 | LLM provider for classifier | Start with OpenAI `gpt-4o-mini` (cheap, fast). Add Ollama local path behind env flag for air-gapped use. |
| D2 | Quantum backend | `qiskit-aer` local simulator is fine for now. Real IBM Quantum backend can be swapped via `IBMQ_TOKEN` env var. |
| D3 | Clock tick frequency | Recompute every 60 s. Debounce signals: only trigger recompute if aggregate distribution shifts by > 2%. |
| D4 | Auth model | API-key-to-JWT for the MVP (no user accounts). Add OAuth2 in Phase 7 if needed. |
| D5 | NATS vs direct gRPC streaming | Use direct gRPC streaming for MVP; NATS adds resilience for multi-instance Gateway later. |
| D6 | nothingx-react-components version | Pin to latest stable; audit peer deps against React 18 before install. |

---

## 11. Definition of Done

- All 4 services pass their own test suites in CI.
- `docker compose up` starts the full stack from cold in < 90 s on an M-series Mac.
- The React SPA shows a live-updating doomsday clock within 5 s of page load.
- The clock responds visually within 10 s of a new signal being classified.
- Globe disintegration effect triggers when `minutesToMidnight < 2.0`.
- Lighthouse performance score ≥ 60 on the main page.
