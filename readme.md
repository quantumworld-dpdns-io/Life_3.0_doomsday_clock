# Life 3.0 Doomsday Clock

> 繁體中文 / English

---

## 繁體中文

本專案是一個 **polyglot monorepo**，用來追蹤世界距離 Max Tegmark《Life 3.0》(2017) 中 **12 種 AI 演化情境**的接近程度：從新聞/研究訊號擷取與 LLM 分類，到以量子熵做種子的風險蒙地卡羅引擎，最後呈現在暗黑賽博風格的 Web UI。

### 系統資料流（Data Flow）

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

### 快速開始（Quick Start）

```bash
cp .env.example .env          # 填入 OPENAI_API_KEY 與 JWT_SECRET
make up                       # 啟動所有容器/服務
# 打開 http://localhost:3000
```

### 常用指令（Common Commands）

| 指令 | 說明 |
|---|---|
| `make up` | 啟動所有服務 |
| `make down` | 停止所有服務 |
| `make logs` | 顯示/追蹤 logs |
| `make test` | 執行所有測試 |
| `make lint` | 執行所有 lint |
| `make proto` | 重新產生 gRPC stubs |
| `make migrate` | 執行 DB migrations |

### 重要環境變數（Key Environment Variables）

| 變數 | 服務 | 說明 |
|---|---|---|
| `OPENAI_API_KEY` | intelligence | LLM 分類器使用的 OpenAI key |
| `OLLAMA_BASE_URL` | intelligence | 本機/自架 LLM（可選） |
| `JWT_SECRET` | api-gateway | HS256 簽章密鑰 |
| `ENABLE_NATS` | api-gateway | 啟用 NATS pub/sub（預設 false） |
| `DATABASE_URL` | intelligence | Postgres 連線字串 |
| `IBMQ_TOKEN` | quantum-sim | IBM Quantum 後端 token（可選） |

### 服務埠號（Service Ports）

| 服務 | HTTP | gRPC |
|---|---:|---:|
| intelligence | 8001 | 50051 |
| quantum-sim | 8002 | 50052 |
| risk-engine | 8003 | 50053 |
| api-gateway | 4000 | — |
| web | 3000 | — |
| postgres | 5432 | — |
| nats | 4222 | — |

### Proto 定義（Proto Definitions）

所有跨服務合約定義在 `shared/proto/`。修改任何 `.proto` 後請執行 `make proto` 重新產生 Go / Python / Rust stubs。

### 文件與決策紀錄（Docs & ADRs）

Tegmark 情境參考與架構決策紀錄（ADR）請見 `apps/docs/`。

---

## English

This project is a **polyglot monorepo** that tracks how close the world is to Max Tegmark’s **12 AI-evolution scenarios** from *Life 3.0* (2017). It ingests news/research signals, classifies them with an LLM, feeds them into a Rust Monte Carlo risk engine seeded with quantum entropy, and visualizes the resulting clock state in a dark cyberpunk web UI.

### Data Flow

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

### Quick Start

```bash
cp .env.example .env          # set OPENAI_API_KEY and JWT_SECRET
make up                       # starts all containers/services
# open http://localhost:3000
```

### Common Commands

| Command | Description |
|---|---|
| `make up` | Start all services |
| `make down` | Stop all services |
| `make logs` | Tail logs |
| `make test` | Run all test suites |
| `make lint` | Lint all services |
| `make proto` | Regenerate gRPC stubs |
| `make migrate` | Run DB migrations |

### Key Environment Variables

| Variable | Service | Description |
|---|---|---|
| `OPENAI_API_KEY` | intelligence | OpenAI key for the LLM classifier |
| `OLLAMA_BASE_URL` | intelligence | Optional local/self-hosted LLM |
| `JWT_SECRET` | api-gateway | HS256 signing secret |
| `ENABLE_NATS` | api-gateway | Enable NATS pub/sub (default false) |
| `DATABASE_URL` | intelligence | Postgres connection string |
| `IBMQ_TOKEN` | quantum-sim | Optional IBM Quantum backend token |

### Service Ports

| Service | HTTP | gRPC |
|---|---:|---:|
| intelligence | 8001 | 50051 |
| quantum-sim | 8002 | 50052 |
| risk-engine | 8003 | 50053 |
| api-gateway | 4000 | — |
| web | 3000 | — |
| postgres | 5432 | — |
| nats | 4222 | — |

### Proto Definitions

All inter-service contracts live in `shared/proto/`. After editing any `.proto` file, run `make proto` to regenerate stubs for Go, Python, and Rust.

### Docs & ADRs

See `apps/docs/` for Tegmark scenario references and architecture decision records (ADRs).
