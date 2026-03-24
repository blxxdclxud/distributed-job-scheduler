# Architecture

**Analysis Date:** 2026-03-24

## Pattern Overview

**Overall:** Distributed Job Scheduling System using Message Queue (RabbitMQ) and Round-Robin Load Balancing

**Key Characteristics:**
- Multi-process architecture with centralized Server and distributed Workers
- Asynchronous communication via RabbitMQ message broker
- Priority-based job queuing (High, Mid, Low)
- Round-robin worker selection algorithm for load balancing
- Lua script execution on worker nodes
- Health reporting and task result collection

## Layers

**API Layer:**
- Purpose: Accept job submissions from clients and return job status
- Location: `server/api/`
- Contains: HTTP request handlers, route definitions, response/error utilities
- Depends on: Scheduler layer
- Used by: HTTP clients submitting jobs

**Scheduler Layer:**
- Purpose: Manage available workers and assign tasks based on priority and availability
- Location: `server/scheduler/`
- Contains: Core scheduling logic, worker round-robin queue, priority-based job queues
- Depends on: Messaging layer (RabbitMQ) for worker communication
- Used by: API layer for job assignment

**Messaging Layer:**
- Purpose: Facilitate asynchronous communication between Server and Workers via RabbitMQ
- Location: `server/messaging/` and `worker/messaging/`
- Contains: RabbitMQ connection management, publishers, consumers, exchange/queue setup
- Depends on: RabbitMQ broker
- Used by: Scheduler (send tasks), Executor (receive tasks), HealthReporter (send heartbeats), all result listeners

**Worker Execution Layer:**
- Purpose: Execute jobs (Lua scripts) on worker instances and report results
- Location: `worker/executor/` and `worker/HealthReporter/`
- Contains: Executor (task listener and Lua interpreter), HealthReporter (periodic heartbeat sender)
- Depends on: Messaging layer, Lua runtime
- Used by: Worker main process

**Shared Models Layer:**
- Purpose: Define common data structures used across Server and Workers
- Location: `shared/models/` and `shared/globals/`
- Contains: Job priority/status enums, Worker model, RabbitMQ message types
- Depends on: None
- Used by: All layers

## Data Flow

**Job Submission Flow:**

1. Client sends HTTP POST to `server/api/SubmitJobHandler` with job script and priority
2. Handler extracts request body, calls `Scheduler.EnqueueJob()`
3. Scheduler locks mutex, creates Job object, adds to `JobQueues` map (grouped by priority), returns jobID
4. Handler responds with jobID and status=PENDING
5. Periodically, `Scheduler.AssignTask()` acquires lock and processes jobs
6. Scheduler calls `RoundRobin()` to get next available worker from `AvailableWorkers` queue
7. If worker available, dequeues highest priority job from `JobQueues`
8. Scheduler calls `rabbitClient.SendTaskToWorker()` to publish LuaTask to RabbitMQ exchange `LuaProgramsExchange` with routing key = workerId
9. Job status updated to RUNNING in `AllJobs` map, tracked in `WorkerAssignments` map
10. If send fails, job re-enqueued to JobQueues

**Job Execution Flow:**

1. Worker's `Executor.ListenTasks()` binds exclusive queue to `LuaProgramsExchange` with routing key = workerId
2. Executor consumes task messages from RabbitMQ
3. On message receipt, extracts LuaCode and JobId from message
4. Executor runs `Task()` method which creates Lua state, opens libraries, executes script
5. Returns result (last value on Lua stack) or error
6. Executor publishes TaskReply to `ResultExchange` with routing key "result.{workerId}"
7. Executor also sends ack (HealthReport) to `WorkerStatusExchangeName` with routing key "heartbeat.{workerId}"

**Worker Health Reporting:**

1. HealthReporter periodically sends HealthReport messages every 5 seconds
2. Messages go to `WorkerStatusExchangeName` with routing key "heartbeat.{workerId}"
3. Server's `ListenHeartBeat()` consumes these messages on heartbeat queue
4. Validates worker liveness via timestamp tracking

**Worker Registration Flow:**

1. Worker main process creates Worker model with unique ID
2. Publishes registration message to `RegisterExchange` with routing key "register"
3. Server's `ListenRegister()` consumes registration messages
4. Extracts worker metadata and creates Worker model
5. Calls `Scheduler.RegisterWorker()` to add to AvailableWorkers queue and TotalWorkers list

**Task Result Collection:**

1. Worker publishes TaskReply containing results, workerId, error, jobId
2. Server's `ListenTaskResults()` consumes from ResultExchange
3. Updates job status in Scheduler's AllJobs map based on success/failure
4. Result stored for retrieval via job status API

**State Management:**

- **Scheduler State:** Protected by mutex for concurrent access
  - `AvailableWorkers`: WorkerQueue (FIFO for round-robin)
  - `TotalWorkers`: Slice of all registered workers
  - `Jobs`: JobQueues (map of priority → priority queue)
  - `AllJobs`: Map[jobID] → Job (tracks status and details)
  - `ReceivedJobsCount`: Counter for job ID generation
  - `WorkerAssignments`: Map[workerID] → []jobID (tracks assignments)

- **Worker State:** Maintained per worker instance
  - Active connections, listener goroutines
  - No persistent state between jobs

- **RabbitMQ State:** Managed by broker
  - Durable exchanges for fault tolerance
  - Exclusive queues on workers for isolation
  - Auto-ack enabled on server listeners (at-least-once semantics)

## Key Abstractions

**Scheduler:**
- Purpose: Central orchestrator managing worker availability and job assignment
- Examples: `server/scheduler/scheduler.go`
- Pattern: Singleton with mutex-protected state, implements round-robin selection

**JobQueues:**
- Purpose: Efficiently manage multiple job priority levels
- Examples: `server/scheduler/task_queue.go`
- Pattern: Map-based queue structure where each priority level has dedicated queue

**WorkerQueue:**
- Purpose: Maintain list of available workers in round-robin order
- Examples: `server/scheduler/worker_queue.go`
- Pattern: FIFO queue wrapper around golang-collections/queue

**RabbitMQ Publisher:**
- Purpose: Standardized message publishing to exchanges
- Examples: `worker/messaging/Rabbit.go`
- Pattern: Interface-based (Publisher interface in HealthReporter)

**Executor:**
- Purpose: Listen for tasks and execute Lua code with result reporting
- Examples: `worker/executor/Executor.go`
- Pattern: Goroutine-based infinite loop consuming RabbitMQ messages

**HealthReporter:**
- Purpose: Periodic health status reporting
- Examples: `worker/HealthReporter/HealthReporter.go`
- Pattern: Ticker-based periodic task with channel-based timeout

## Entry Points

**Server Entry Point:**
- Location: `cmd/server/main.go`
- Triggers: Binary execution with optional `-rmq` flag for RabbitMQ URI
- Responsibilities:
  - Parse command-line arguments
  - Initialize logger (development mode)
  - Call `server.RunServer()`

**Server RunServer:**
- Location: `server/server.go`
- Triggers: Called from main
- Responsibilities:
  - Initialize Scheduler
  - Create API Handler with Scheduler reference
  - Register HTTP routes via `api.RegisterRoutes()`
  - Start HTTP server on port 8080

**Worker Entry Point:**
- Location: `cmd/worker/main.go`
- Triggers: Binary execution with optional `-rmq` flag
- Responsibilities:
  - Parse RabbitMQ connection string
  - Establish RabbitMQ connection
  - Initialize logger (local mode with slog)
  - Create Worker model with unique ID
  - Publish worker registration message
  - Call `worker.Start()` to spawn Executor and HealthReporter goroutines
  - Wait for shutdown signal (SIGINT/SIGTERM)

## Error Handling

**Strategy:** Multi-level error handling with logging, no panic recovery

**Patterns:**
- Failed job sends: Re-enqueue job to JobQueues, log error
- RabbitMQ connection errors: Panic on fatal errors (connection setup)
- Lua execution errors: Return error in TaskReply, continue listening
- HTTP request parsing: Return HTTP 400/500 error responses
- Channel operations: Log errors, continue operation (non-blocking pattern)

## Cross-Cutting Concerns

**Logging:**
- Framework: Zap logger (`pkg/logger/logger.go`) on server, slog on worker
- Pattern: Global logger initialized at startup with environment-based configuration
- Server: Development mode with color-coded levels
- Worker: Slog with environment-based JSON/text formatting

**Validation:**
- API layer validates request format (JSON parsing)
- Priority values validated against JobPriorities constant
- RabbitMQ exchanges/queues declared with existence checks

**Authentication:**
- RabbitMQ connection: Basic auth via AMQP URI (guest:guest in defaults)
- No per-message authentication
- No API client authentication

---

*Architecture analysis: 2026-03-24*
