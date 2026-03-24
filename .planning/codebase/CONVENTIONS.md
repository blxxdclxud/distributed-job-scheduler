# Coding Conventions

**Analysis Date:** 2026-03-24

## Naming Patterns

**Files:**
- PascalCase for files with exported types (e.g., `Executor.go`, `HealthReporter.go`)
- snake_case for utility and internal files (e.g., `api_scheduler_test.go`, `job_queues.go`)
- PascalCase for test files using action naming (e.g., `HealthCheck_test.go`)

**Functions:**
- PascalCase for exported functions (e.g., `NewScheduler()`, `SubmitJobHandler()`, `RegisterRoutes()`)
- PascalCase for methods (e.g., `EnqueueJob()`, `ListenHeartBeat()`, `SendHealthChecks()`)
- Utility functions use PascalCase with descriptive prefixes (e.g., `ErrorResponse()`, `ResponseJson()`, `SetRabbitClient()`)

**Variables:**
- camelCase for local variables and parameters (e.g., `jobID`, `workerId`, `jobRequest`)
- camelCase for interface receivers (e.g., `h *Handler`, `s *Scheduler`, `r *Rabbit`, `e *Executor`)
- ALL_CAPS for package-level constants (e.g., `SendJobTimeout`)
- PascalCase for struct fields (e.g., `JobID`, `Priority`, `Script`, `AvailableWorkers`)

**Types:**
- PascalCase for struct types (e.g., `Handler`, `Scheduler`, `Job`, `Worker`, `Rabbit`)
- PascalCase for interface types (e.g., `Executor`, `HealthReporter`, `Publisher`)
- PascalCase for type aliases (e.g., `JobPriority`, `JobStatus`)

**Constants:**
- ALL_CAPS for global constants (e.g., `HighPriority`, `LowPriority`, `StatusPending`, `StatusRunning`)
- Descriptive naming with domain context (e.g., `LuaProgramsExchange`, `WorkerStatusExchangeName`, `ResultExchange`)

## Code Style

**Formatting:**
- Go's standard `gofmt` formatting (implicit - no custom formatter detected)
- Consistent 2-3 space indentation following Go conventions
- Line length respects Go's informal 80-100 character guideline

**Linting:**
- No custom `.eslintrc`, `.prettierrc`, or `biome.json` detected
- Standard Go tooling implied (go fmt, go vet)

**Brace Placement:**
- Idiomatic Go style: opening braces on same line as declaration
- Example from `server/scheduler/scheduler.go`:
```go
func (s *Scheduler) NewScheduler() *Scheduler {
    return &Scheduler{
        AvailableWorkers:  *NewWorkerQueue(),
        Jobs:              *NewJobQueues(),
        ReceivedJobsCount: 0,
    }
}
```

## Import Organization

**Order:**
1. Standard library imports (`encoding/json`, `fmt`, `context`, `net/http`, etc.)
2. External dependency imports (`github.com/` packages)
3. Internal module imports (gitlab path: `gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/`)

**Path Aliases:**
- Use aliases to avoid conflicts with multiple `models` packages: `sharedModels "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"`
- Use aliases for clarity when importing subpackages: `Executor2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/executor"`
- Aliases use PascalCase with number suffix for disambiguation (e.g., `Rabbit2`, `HealthReporter2`)

**Example from `server/api/handler.go`:**
```go
import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/models"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/scheduler"
	sharedModels "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)
```

## Error Handling

**Patterns:**
- Immediate check after operations: `if err != nil { ... }`
- Early returns from functions: `if err != nil { return ... }`
- Logging errors with context using `logger.Error()` or `log.Printf()`
- Custom error helper function: `failOnError(err, "message")` in tests that panics on error
- Wrapped errors using `fmt.Errorf()` with context: `return nil, fmt.Errorf("lua execution error: %w", err)`

**Example from `server/api/handler.go`:**
```go
if err := json.NewDecoder(r.Body).Decode(&jobRequest); err != nil {
    ErrorResponse(w, http.StatusBadRequest, "invalid request format")
    return
}
```

**Example from `worker/executor/Executor.go`:**
```go
if err := lua.DoString(l, body); err != nil {
    return nil, fmt.Errorf("lua execution error: %w", err)
}
```

## Logging

**Framework:**
- Primary: `go.uber.org/zap` (structured logging)
- Secondary: `log/slog` (standard library structured logging) in worker components
- Fallback: `log` standard library package for basic logging
- Console output: `fmt.Println()`, `fmt.Printf()` for debugging in tests

**Patterns:**

Zap logging with fields (primary pattern):
```go
logger.Info("Received job from a client", zap.Int("ID", jobID))
logger.Error("Client's request failed", zap.Int("HTTP code", status), zap.String("Error", msg))
```

Slog logging with key-value pairs (worker pattern):
```go
h.log.Info("HealthReporter SendHealthChecks")
h.log.Error("HealthReporter SendHealthChecks", "err", err)
```

- Log when jobs are received and processed
- Log errors at INFO level when client requests fail (with HTTP code)
- Log task assignments and reassignments at DEBUG level
- Use descriptive messages with contextual field names

## Comments

**When to Comment:**
- Every exported function and type should have a comment explaining its purpose
- Comments for non-obvious logic or important decisions
- Comments clarifying parameter meanings in inline declarations (AMQP channel parameters)

**JSDoc/TSDoc:**
- Go does not use JSDoc, instead uses single-line comments before declarations
- Comment format: `// FunctionName brief description`
- Example from `server/api/handler.go`:
```go
// Handler stores Scheduler instance as field, that allows to pass new arrived jobs to it
type Handler struct {
    Scheduler *scheduler.Scheduler
}

// SubmitJobHandler is handler that accepts the job submitted by a client.
// It passes the job to the Scheduler in case of successful
func (h *Handler) SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
```

**Inline Comments:**
- Parameter descriptions in AMQP declarations use inline comments for clarity:
```go
ch.QueueDeclare(
    "",    // name
    false, // durable
    false, // delete when unused
    true,  // exclusive
)
```

## Function Design

**Size:**
- Functions are typically 20-40 lines for business logic
- Larger functions (60+ lines) used for complex message handling loops
- Utility functions kept concise (5-15 lines)

**Parameters:**
- Use receiver pattern for methods: `(h *Handler)`, `(s *Scheduler)`
- Simple parameters preferred; complex operations wrapped in structs
- Context parameters passed for timeout control: `ctx context.Context`

**Return Values:**
- Error as last return value: `(int, error)` or `(*Scheduler, error)`
- Multiple return values used for status + error pairs
- Tuples with bool for optional returns: `if job, ok := queue.Get(); ok {`

## Module Design

**Exports:**
- Exported functions start with capital letter (Go convention)
- Exported types and constants use PascalCase and ALL_CAPS respectively
- Private functions and types use lowercase

**Barrel Files:**
- No barrel/index.ts pattern observed (Go-specific)
- Packages are organized by functionality, not re-exported through a central module

**Package Naming:**
- Package names match directory names: `package api`, `package scheduler`, `package executor`
- Simple, single-word package names preferred
- Avoid generic names; use domain-specific names (e.g., `messaging`, `HealthReporter`)

## Struct Field Tagging

**JSON Tags:**
- Struct fields include JSON tags for HTTP request/response marshaling
- snake_case for JSON field names with optional `,omitempty` for optional fields
- Example from `server/models/api.go`:
```go
type JobRequest struct {
    Script   string `json:"script"`
    Priority int    `json:"priority,omitempty"`
}
```

## Interface Design

**Usage:**
- Interfaces used for dependency injection and abstraction
- Small, focused interfaces (1-2 methods)
- Example from `shared/models/worker.go`:
```go
type Executor interface {
    ListenTasks(workerId string)
}

type HealthReporter interface {
    SendHealthChecks(workerId string)
}
```

**Nil Checks:**
- Defensive checks for nil returns: `if worker != nil { ... }`
- Used before dereferencing pointers

---

*Convention analysis: 2026-03-24*
