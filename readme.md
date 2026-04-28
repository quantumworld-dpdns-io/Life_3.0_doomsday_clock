# Life 3.0 Doomsday Clock

> 繁體中文 / English

---

## 繁體中文

本專案是一個 **polyglot monorepo**，用來追蹤世界距離 Max Tegmark《Life 3.0》(2017) 中 **12 種 AI 演化情境**的接近程度：從新聞/研究訊號擷取與 LLM 分類，到以量子熵做種子的風險蒙地卡羅引擎，最後呈現在暗黑賽博風格的 Web UI。

### 目前進度（Current Progress）

> 狀態快照：2026-04-28，以目前工作樹中存在的檔案/目錄為準。

| 區域 | 狀態 | 備註 |
|---|---|---|
| Shared proto / scripts | 已有檔案 | `shared/proto/` 與 `shared/scripts/gen_proto.sh` 存在；目前 gateway MVP 走 HTTP upstream，gRPC stub 產生仍可作為後續強化。 |
| Intelligence service | 已實作並通過測試 | Python app、scraper/classifier、migration、Dockerfile、tests 皆存在；`uv run --extra dev pytest` 20 passed，ruff passed。 |
| Quantum simulator | 已實作並通過測試 | QASM circuit、runner、API/gRPC、Dockerfile、tests 皆存在；`uv run --extra dev pytest` 18 passed，ruff passed。 |
| Risk engine | 已實作並通過測試 | Rust source、config、build script、Dockerfile 存在；`cargo test` 6 passed，HTTP `/clock` 可供 gateway 讀取。 |
| API gateway | MVP 已實作 | Go `go.mod`、Dockerfile、login/JWT、GraphQL-compatible handler、upstream client 與 tests 已存在；`go test ./...` passed。 |
| Web frontend | MVP 已實作 | Vite React app、CRT overlay、Three.js globe、clock、scenario panel、signal feed、fallback data、tests 與 Dockerfile 已存在；build/lint/test passed。 |
| Compose / Make / CI | 已實作 | `docker-compose.yaml`、`Makefile`、`.github/workflows/ci.yaml`、lint/test scripts 存在；`docker compose config` passed，完整 build/up 尚待最終驗證。 |

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

完整 stack 啟動流程如下；目前已完成 MVP 實作，仍建議在部署前跑完整 Docker Compose build/up 驗證。

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

### Supabase 資料庫設定

若使用 Supabase shared transaction pooler，請在 `.env` 或部署平台環境變數設定：

```bash
DATABASE_URL=postgresql://postgres.meslvfoeoduhbsawgawz:YOUR_URL_ENCODED_PASSWORD@aws-1-ap-southeast-2.pooler.supabase.com:6543/postgres?sslmode=require
```

密碼如果包含 `@`, `:`, `/`, `?`, `#`, `%` 等字元，需先 URL encode。不要將真實密碼提交到 Git。

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

### Current Progress

> Snapshot: 2026-04-28, based on files/directories present in the current working tree.

| Area | Status | Notes |
|---|---|---|
| Shared proto / scripts | Files present | `shared/proto/` and `shared/scripts/gen_proto.sh` exist; the gateway MVP currently uses HTTP upstreams, with generated gRPC stubs left as a hardening path. |
| Intelligence service | Implemented and tested | Python app, scraper/classifier modules, migration, Dockerfile, and tests are present; `uv run --extra dev pytest` passed 20 tests and ruff passed. |
| Quantum simulator | Implemented and tested | QASM circuit, runner, API/gRPC modules, Dockerfile, and tests are present; `uv run --extra dev pytest` passed 18 tests and ruff passed. |
| Risk engine | Implemented and tested | Rust source, config, build script, and Dockerfile are present; `cargo test` passed 6 tests and HTTP `/clock` supports gateway reads. |
| API gateway | MVP implemented | Go `go.mod`, Dockerfile, login/JWT, GraphQL-compatible handler, upstream client, and tests are present; `go test ./...` passed. |
| Web frontend | MVP implemented | Vite React app, CRT overlay, Three.js globe, clock, scenario panel, signal feed, fallback data, tests, and Dockerfile are present; build/lint/test passed. |
| Compose / Make / CI | Implemented | `docker-compose.yaml`, `Makefile`, `.github/workflows/ci.yaml`, and lint/test scripts exist; `docker compose config` passed, with full build/up still pending final verification. |

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

The full-stack startup flow is below. The MVP implementation is present; run a full Docker Compose build/up before treating it as deployment-ready.

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

### Supabase Database

For the Supabase shared transaction pooler, set this in `.env` or your deployment environment:

```bash
DATABASE_URL=postgresql://postgres.meslvfoeoduhbsawgawz:YOUR_URL_ENCODED_PASSWORD@aws-1-ap-southeast-2.pooler.supabase.com:6543/postgres?sslmode=require
```

URL-encode the password if it contains characters such as `@`, `:`, `/`, `?`, `#`, or `%`. Do not commit the real password.

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
