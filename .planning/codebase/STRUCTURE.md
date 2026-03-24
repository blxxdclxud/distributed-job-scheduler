# Codebase Structure

**Analysis Date:** 2026-03-24

## Directory Layout

```
DistributedJobScheduling/
├── cmd/                         # Application entry points
│   ├── server/                  # Server binary entrypoint
│   │   └── main.go
│   └── worker/                  # Worker binary entrypoint
│       └── main.go
├── server/                      # Server application (scheduler, API, messaging)
│   ├── api/                     # HTTP API handlers and routing
│   │   ├── handler.go          # Handler struct with API endpoints
│   │   ├── routes.go           # Route registration
│   │   ├── utils.go            # Response/error formatting
│   │   └── README.md
│   ├── scheduler/               # Job scheduling and worker management
│   │   ├── scheduler.go        # Core scheduler logic
│   │   ├── task_queue.go       # Priority-based job queuing
│   │   └── worker_queue.go     # Round-robin worker queue
│   ├── messaging/               # RabbitMQ communication for server
│   │   ├── rabbit.go           # RabbitMQ client initialization
│   │   ├── ListenHeartBeat.go  # Worker health monitoring
│   │   ├── ListenTaskResults.go # Task result collection
│   │   ├── ListenRegister.go   # Worker registration handling
│   │   ├── SendTaskToWorker.go # Task distribution
│   │   └── tests/              # Messaging layer tests
│   ├── models/                  # Server-specific data models
│   │   ├── job.go              # Job model (wraps shared Job)
│   │   └── api.go              # API request/response models
│   ├── server.go               # Server initialization and HTTP setup
│   ├── tests/                  # Server integration tests
│   │   └── api_scheduler_test.go
│   ├── config/                 # Configuration (empty, placeholder)
│   └── utils/                  # Server utilities
│       └── InitRabbit/         # RabbitMQ queue initialization
├── worker/                      # Worker application (execution, health reporting)
│   ├── executor/                # Lua script execution
│   │   └── Executor.go         # Task listener and Lua interpreter
│   ├── HealthReporter/          # Worker health status reporting
│   │   └── HealthReporter.go   # Periodic heartbeat sender
│   ├── messaging/               # RabbitMQ communication for worker
│   │   └── Rabbit.go           # RabbitMQ publisher client
│   ├── tests/                  # Worker unit tests
│   │   └── HealthCheck_test.go
│   ├── config/                 # Configuration (empty, placeholder)
│   └── utils/                  # Worker utilities (empty)
├── shared/                      # Shared code between server and worker
│   ├── models/                  # Shared data structures
│   │   ├── job.go              # Job priority and status enums
│   │   ├── worker.go           # Worker model and interfaces
│   │   └── Rabbit/             # RabbitMQ message types
│   │       ├── Registration.go
│   │       ├── hearthBeat.go
│   │       ├── LuaTask.go
│   │       └── TaskReply.go
│   └── globals/                 # Global constants
│       └── globals.go          # RabbitMQ exchange/queue names
├── pkg/                         # Shared libraries and utilities
│   └── logger/                  # Logging utilities
│       └── logger.go           # Zap logger configuration and wrappers
├── deployments/                 # Deployment configurations
│   └── docker-compose.yml      # Local development environment
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies lock
├── Dockerfile                   # Container image definition
└── README.md                    # Project documentation
```

## Directory Purposes

**cmd/**
- Purpose: Binary entry points for server and worker applications
- Contains: main() functions only
- Key files: `cmd/server/main.go`, `cmd/worker/main.go`

**server/**
- Purpose: Central job scheduling server application
- Contains: API handlers, scheduling logic, message queue management
- Key files: `server/server.go` (initialization), `server/scheduler/scheduler.go` (core logic)

**server/api/**
- Purpose: HTTP request handling and routing
- Contains: Handler methods for /submit_job and /status/{id} endpoints
- Key files: `server/api/handler.go` (endpoints), `server/api/routes.go` (route registration)

**server/scheduler/**
- Purpose: Job queue management and worker assignment algorithms
- Contains: Scheduler state, round-robin selection, priority queuing
- Key files: `server/scheduler/scheduler.go` (Scheduler struct), `server/scheduler/task_queue.go` (JobQueues)

**server/messaging/**
- Purpose: RabbitMQ integration for server-to-worker communication
- Contains: Message listeners (tasks, heartbeats, results, registration), task distribution
- Key files: `server/messaging/rabbit.go` (client init), `server/messaging/SendTaskToWorker.go` (task dispatch)

**server/models/**
- Purpose: Server-specific data models not shared with workers
- Contains: Job struct (extends shared), JobRequest/JobResponse for API
- Key files: `server/models/job.go`

**worker/**
- Purpose: Distributed worker application executing Lua jobs
- Contains: Task execution engine, health reporting, RabbitMQ communication
- Key files: `worker/executor/Executor.go` (task listener), `worker/HealthReporter/HealthReporter.go` (health checks)

**worker/executor/**
- Purpose: Listen for and execute Lua scripts
- Contains: RabbitMQ message consumer, Lua state management, result publishing
- Key files: `worker/executor/Executor.go` (implements Executor interface)

**worker/HealthReporter/**
- Purpose: Periodic worker availability reporting
- Contains: Ticker-based heartbeat sender, health status messages
- Key files: `worker/HealthReporter/HealthReporter.go` (implements HealthReporter interface)

**shared/models/**
- Purpose: Common data types used by both server and worker
- Contains: Job enums (priority, status), Worker model, RabbitMQ message types
- Key files: `shared/models/job.go` (enums), `shared/models/worker.go` (Worker struct)

**shared/models/Rabbit/**
- Purpose: RabbitMQ message type definitions
- Contains: Message wrapper types with JSON marshaling support
- Key files: `LuaTask.go` (task payload), `TaskReply.go` (result payload), `hearthBeat.go` (health status)

**shared/globals/**
- Purpose: Application-wide constants
- Contains: RabbitMQ exchange and queue names
- Key files: `shared/globals/globals.go` (defines LuaProgramsExchange, ResultExchange, etc)

**pkg/logger/**
- Purpose: Centralized logging configuration and utilities
- Contains: Zap logger wrapper with helper functions (Info, Error, Debug, etc)
- Key files: `pkg/logger/logger.go` (global logger init and convenience functions)

## Key File Locations

**Entry Points:**
- `cmd/server/main.go`: Server initialization, parses -rmq flag, calls RunServer()
- `cmd/worker/main.go`: Worker initialization, registers with server, starts execution loop
- `server/server.go`: RunServer() initializes API and scheduler on port 8080

**Configuration:**
- `shared/globals/globals.go`: RabbitMQ exchange/queue names (centralized)
- `deployments/docker-compose.yml`: Local RabbitMQ broker setup
- `.planning/codebase/`: Analysis documents (generated)

**Core Logic:**
- `server/scheduler/scheduler.go`: Scheduler struct with EnqueueJob, AssignTask, RoundRobin methods
- `server/scheduler/task_queue.go`: JobQueues implementation (map of priority → queue)
- `server/scheduler/worker_queue.go`: WorkerQueue implementation (round-robin FIFO)
- `worker/executor/Executor.go`: Task listening and Lua execution (ListenTasks, Task methods)
- `server/api/handler.go`: API endpoints SubmitJobHandler, GetJobStatusHandler

**Testing:**
- `server/tests/api_scheduler_test.go`: Scheduler and API integration tests (48 lines)
- `worker/tests/HealthCheck_test.go`: Worker health reporting tests (242 lines)
- `server/messaging/tests/RabbitFunctions_test.go`: RabbitMQ integration tests (59 lines)

## Naming Conventions

**Files:**
- PascalCase for exported types: `Executor.go`, `HealthReporter.go`, `Rabbit.go`
- snake_case for unexported/utility files: `task_queue.go`, `worker_queue.go`, `api.go`
- Tests: `*_test.go` suffix in same package as code under test
- Init utilities: `Init{Name}.go` in utils subdirectories

**Directories:**
- Lower case with hyphens where applicable
- Functional grouping (api/, scheduler/, messaging/) for related code
- PascalCase for packages with types: `HealthReporter/`, `Rabbit/`
- Shared code in `shared/` hierarchy for cross-process code

**Functions:**
- PascalCase for exported functions: `NewScheduler()`, `EnqueueJob()`, `ListenTasks()`
- camelCase for unexported functions: `roundRobin()`, `assignTaskToWorkerUtil()`
- Handler methods follow `{Action}{Entity}Handler` pattern: `SubmitJobHandler`, `GetJobStatusHandler`

**Variables:**
- Short camelCase in local scopes: `job`, `worker`, `msgs`, `err`
- Descriptive camelCase for package-level: `ReceivedJobsCount`, `AvailableWorkers`
- All caps for constants: `SendJobTimeout`, `LuaProgramsExchange`, `StatusRunning`

**Types:**
- PascalCase for struct names: `Scheduler`, `Job`, `Worker`, `Handler`
- PascalCase for interface names: `Executor`, `HealthReporter`, `Publisher`
- Enum type PascalCase: `JobPriority`, `JobStatus`

## Where to Add New Code

**New Feature (e.g., job timeout handling):**
- Primary code: `server/scheduler/scheduler.go` (add timeout logic to job tracking)
- Supporting code: `server/models/job.go` (add timeout field to Job struct)
- Tests: `server/tests/api_scheduler_test.go` (add test cases)
- Shared models: `shared/models/job.go` (if needed for worker coordination)

**New Component/Module (e.g., persistence layer):**
- Implementation: `server/storage/` (create new directory)
- Integrate: Modify `server/server.go` to initialize and inject dependency
- Models: Add models to `server/models/` if server-specific, `shared/models/` if shared
- Tests: Create `server/storage/storage_test.go`

**Utilities:**
- Shared helpers: `pkg/{functionality}/` (e.g., `pkg/utils/`, `pkg/validation/`)
- Server-specific utilities: `server/utils/{functionality}/`
- Worker-specific utilities: `worker/utils/{functionality}/`
- Example: `server/utils/InitRabbit/` for RabbitMQ setup helpers

**New API Endpoint:**
- Handler: Add method to `Handler` struct in `server/api/handler.go`
- Route: Register in `server/api/routes.go` with `router.HandleFunc()`
- Models: Add request/response types to `server/models/api.go`
- Tests: Add test case to `server/tests/api_scheduler_test.go`

**New Worker Capability (e.g., Python execution):**
- Implementation: Create `worker/python_executor/` directory parallel to executor/
- Implement: Executor interface (ListenTasks method) from `shared/models/worker.go`
- Integration: Modify Worker struct initialization in `shared/models/worker.go`
- Tests: Create `worker/python_executor/python_executor_test.go`

## Special Directories

**deployments/:**
- Purpose: Contains production/development deployment configurations
- Generated: No
- Committed: Yes
- Contents: docker-compose.yml for local RabbitMQ

**.planning/codebase/:**
- Purpose: Analysis documents for architecture, structure, conventions, testing
- Generated: Yes (by GSD map-codebase command)
- Committed: Yes
- Contents: ARCHITECTURE.md, STRUCTURE.md, CONVENTIONS.md, TESTING.md, CONCERNS.md

**server/config/ and worker/config/:**
- Purpose: Configuration file storage (currently empty placeholders)
- Generated: No
- Committed: No (likely ignored)
- Intended use: Environment-specific configuration files

**shared/:**
- Purpose: Code shared between server and worker binaries
- Generated: No
- Committed: Yes
- Constraint: Must not import from server/ or worker/ packages

---

*Structure analysis: 2026-03-24*
