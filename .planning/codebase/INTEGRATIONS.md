# External Integrations

**Analysis Date:** 2026-03-24

## APIs & External Services

**Job Submission API:**
- HTTP REST API exposed at port 8080
  - SDK/Client: Gorilla Mux router
  - Endpoints: `POST /jobs` (submit job), `GET /jobs/{id}` (get job status)
  - Location: `server/api/handler.go`, `server/api/routes.go`

**Task Execution:**
- Lua script engine - Executes arbitrary Lua code provided in job submissions
  - SDK/Client: Shopify go-lua
  - Location: `worker/executor/Executor.go`

## Message Broker

**RabbitMQ:**
- AMQP 0.9.1 protocol
- Connection: Via `-rmq` CLI flag (default: `amqp://guest:guest@localhost:5672/`)
- Client: `github.com/rabbitmq/amqp091-go` v1.10.0
- Credentials: Hardcoded as `guest:guest` in Docker Compose and CLI defaults

**Exchanges:**
- `lua_programs` (direct exchange) - Distributes Lua tasks from server to specific workers
  - Location: `server/messaging/rabbit.go`, `worker/executor/Executor.go`
- `worker_status` (topic exchange) - Receives worker heartbeat and status reports
  - Location: `server/messaging/rabbit.go`, `server/messaging/ListenHeartBeat.go`
- `task_results` (topic exchange) - Receives execution results from workers
  - Location: `server/messaging/rabbit.go`, `server/messaging/ListenTaskResults.go`
- `register` (direct exchange) - Handles worker registration on startup
  - Location: `server/messaging/rabbit.go`, `cmd/worker/main.go`

**Queues:**
- Worker registration queue: Declared in `server/utils/InitRabbit/InitRegister.go`
- Heartbeat queue: Declared in `server/utils/InitRabbit/InitHeartBeat.go`
- Task result queue: Declared in `server/utils/InitRabbit/InitTaskResult.go`

## Data Storage

**Databases:** None - Project uses in-memory state management

**File Storage:** Not applicable - No file storage integrations detected

**Caching:** None - No caching layer detected

## Worker Communication

**Worker Registry:**
- Workers register themselves on startup via RabbitMQ `register` exchange
- Location: `cmd/worker/main.go` (publishes registration message)
- Registration message format: JSON containing worker ID (UUID)

**Heartbeat:**
- Workers send periodic heartbeat messages to `worker_status` exchange
- Routing key: `heartbeat.{workerId}`
- Message type: `HealthReport` with `WorkerId` and `TimeStamp`
- Location: `worker/HealthReporter/HealthReporter.go`

**Task Distribution:**
- Server sends Lua tasks via `lua_programs` exchange
- Routing key: Worker ID (direct targeting)
- Message type: `LuaTask` containing job ID and Lua code
- Location: `server/scheduler/scheduler.go` (sends), `worker/executor/Executor.go` (receives)

**Result Collection:**
- Workers send execution results via `task_results` exchange
- Routing key: `result.{workerId}`
- Message type: `TaskReply` containing job ID, results, worker ID, and errors
- Location: `worker/executor/Executor.go` (sends), `server/messaging/ListenTaskResults.go` (receives)

## Authentication & Identity

**Auth Provider:** Custom implementation

**Worker Identification:**
- Uses UUID generated per worker instance
- Unique ID assigned at startup: `id := uuid.New().String()`
- Location: `cmd/worker/main.go`

**RabbitMQ Auth:**
- Basic authentication with hardcoded credentials
- Default: `guest:guest`
- Configurable via connection string in `-rmq` flag

## Monitoring & Observability

**Error Tracking:** None detected

**Logs:**
- Approach: Uber Zap structured logging
- Levels: DEBUG (development), INFO/ERROR (production)
- Location: `pkg/logger/logger.go`
- Time format: DD-MM-YYYY HH:MM:SS
- Worker logs: slog package (standard library JSON/text handlers)
- Location: `cmd/worker/main.go`, `worker/executor/Executor.go`

## Containerization & Deployment

**Hosting:** Docker containers (local development or Kubernetes-ready)

**CI/CD Pipeline:**
- GitLab CI with linting stage
- Config: `.gitlab-ci.yml`
- Linter: golangci-lint
- No automated deployment configured

**Docker Compose (Local Development):**
- Service: `rabbitmq:3-management` - Message broker
- Service: `host` - Server (port 8080 exposed)
- Service: `worker1`, `worker2` - Worker instances
- Network: Internal network for service-to-service communication
- External network for client access to server API
- Location: `deployments/docker-compose.yml`

## Environment Configuration

**Required env vars/flags:**
- `-rmq` - RabbitMQ connection URL (default: `amqp://guest:guest@localhost:5672/`)

**Secrets location:**
- RabbitMQ credentials embedded in code and Docker Compose
- Note: `.env` files are not used; credentials passed via CLI flags

## Webhooks & Callbacks

**Incoming:**
- HTTP POST `/jobs` - Job submission endpoint
- HTTP GET `/jobs/{id}` - Status query endpoint

**Outgoing:**
- None - Results are pulled from RabbitMQ by server

---

*Integration audit: 2026-03-24*
