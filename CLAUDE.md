# Life 3.0 Doomsday Clock

A polyglot monorepo that tracks the world's proximity to Max Tegmark's 12 AI-evolution scenarios from *Life 3.0* (2017). News and research signals are classified by an LLM, fed into a probabilistic Rust engine seeded with genuine quantum randomness, and displayed on a dark cyberpunk React SPA.

## Data Flow

```
Reuters / arXiv / HN
        │
        ▼
[Python] services/intelligence  — scraper + LLM classifier (FastAPI + LangChain)
        │  ScenarioSignal (gRPC)
        ▼
[Rust]  services/risk-engine    — Monte Carlo probability engine (Tonic gRPC)
        ↑  EntropyResponse (gRPC)
[QASM]  services/quantum-sim    — Hadamard-gate entropy source (Qiskit + FastAPI)
        │  ClockState (gRPC)
        ▼
[Go]    services/api-gateway    — GraphQL + JWT gateway (gqlgen)
        │  GraphQL / WebSocket
        ▼
[React] apps/web                — Dark UI (Three.js globe, CRT clock, scenario panel)
```

## Quick Start

```bash
cp .env.example .env          # fill in OPENAI_API_KEY and JWT_SECRET
make up                       # starts all 7 containers
# open http://localhost:3000
```

## Common Commands

| Command | Description |
|---------|-------------|
| `make up` | Start all services |
| `make down` | Stop all services |
| `make logs` | Tail logs |
| `make test` | Run all test suites |
| `make lint` | Lint all services |
| `make proto` | Regenerate gRPC stubs |
| `make migrate` | Run DB migrations |

## Key Environment Variables

| Variable | Service | Description |
|----------|---------|-------------|
| `OPENAI_API_KEY` | intelligence | OpenAI key for LLM classifier |
| `OLLAMA_BASE_URL` | intelligence | Alternative local LLM (optional) |
| `JWT_SECRET` | api-gateway | HS256 signing secret |
| `ENABLE_NATS` | api-gateway | Enable NATS pub/sub (default false) |
| `DATABASE_URL` | intelligence | Postgres connection string |
| `IBMQ_TOKEN` | quantum-sim | Real IBM Quantum backend (optional) |

## Service Ports

| Service | HTTP | gRPC |
|---------|------|------|
| intelligence | 8001 | 50051 |
| quantum-sim | 8002 | 50052 |
| risk-engine | 8003 | 50053 |
| api-gateway | 4000 | — |
| web | 3000 | — |
| postgres | 5432 | — |
| nats | 4222 | — |

## Proto Definitions

All inter-service contracts live in `shared/proto/`. After editing any `.proto` file run `make proto` to regenerate stubs for Go, Python, and Rust.

## Architecture Decisions

See `apps/docs/` for Tegmark scenario reference and ADRs.
