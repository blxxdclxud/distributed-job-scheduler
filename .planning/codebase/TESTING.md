# Testing Patterns

**Analysis Date:** 2026-03-24

## Test Framework

**Runner:**
- Go's standard `testing` package
- No custom test config file (`testing.json`, `jest.config.js`, etc.)
- Tests run with: `go test ./...`

**Assertion Library:**
- Standard Go error checks: `if err != nil { ... }`
- No external assertion library (testify, goconvey, etc.)
- Manual assertions via error comparisons and value checks

**Run Commands:**
```bash
go test ./...                 # Run all tests in all packages
go test ./path/to/package     # Run tests in specific package
go test -v ./...              # Verbose output with individual test names
go test -run TestName ./...   # Run specific test by pattern
```

## Test File Organization

**Location:**
- Tests are co-located with source code in same package
- Pattern: test files live in same directory as implementation
- `tests/` subdirectory pattern also used (see `server/tests/`, `worker/tests/`, `server/messaging/tests/`)
- Tests in `tests/` subdirectory are preferred for integration/system tests

**Naming:**
- Pattern: `*_test.go`
- Examples: `api_scheduler_test.go`, `HealthCheck_test.go`, `RabbitFunctions_test.go`

**Structure:**
```
server/
├── tests/
│   └── api_scheduler_test.go
├── api/
│   ├── handler.go
│   └── utils.go
├── scheduler/
│   └── scheduler.go
└── messaging/
    └── tests/
        └── RabbitFunctions_test.go

worker/
├── tests/
│   └── HealthCheck_test.go
├── executor/
│   └── Executor.go
└── HealthReporter/
    └── HealthReporter.go
```

## Test Structure

**Suite Organization:**

Standard Go test function pattern from `server/tests/api_scheduler_test.go`:
```go
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
	"net/http"
	"testing"
)

func TestAPIAndScheduler(t *testing.T) {
	// Setup
	testData := []job{
		{
			priority: models.LowPriority,
			script:   `local start = os.time(); while os.time() - start < 2 do end; ...`,
		},
	}

	// Run
	server.RunServer()

	for _, j := range testData {
		buff, _ := json.Marshal(j)
		resp, err := http.Post(
			"http://localhost:8080",
			"application/json",
			bytes.NewReader(buff),
		)
		fmt.Println(resp, err)
	}
	// Assertions implicit via error checks and output verification
}
```

**Patterns:**

Setup pattern:
- Test data defined as slice of structs in test function
- Server/component instantiation at start of test
- No separate setup helper functions (all inline)

Teardown pattern:
- Deferred cleanup: `defer conn.Close()`, `defer cancel()`
- Channel signaling for shutdown: `sigChan := make(chan os.Signal, 1)`

Assertion pattern:
- No explicit assertions; verification via:
  - Error checking: `if err != nil { ... }`
  - Channel receive with timeout: `select { case result := <-ch: ... }`
  - Log output inspection (tests output to console)
  - Response code checks implicit in HTTP response objects

## Mocking

**Framework:**
- No mocking library detected
- Manual mocking via interface implementation
- Test implementations satisfy interfaces defined in production code

**Patterns:**

Interface-based mocking from tests. For example, `Executor` interface from `shared/models/worker.go`:
```go
type Executor interface {
    ListenTasks(workerId string)
}
```

Tests implement this interface:
```go
type MockExecutor struct{}
func (m *MockExecutor) ListenTasks(workerId string) {
    // Mock implementation
}
```

**What to Mock:**
- External dependencies: RabbitMQ connections, HTTP clients, file I/O
- Components that are slow or have side effects (database calls, network requests)
- Time-dependent operations (use `time.After()` for timeouts in tests)

**What NOT to Mock:**
- Core business logic (schedulers, handlers, executors)
- Data models and structs
- Pure utility functions

## Fixtures and Factories

**Test Data:**

Example from `server/tests/api_scheduler_test.go`:
```go
type job struct {
	priority models.JobPriority
	script   string
}

testData := []job{
	{
		priority: models.LowPriority,
		script:   `local start = os.time(); while os.time() - start < 2 do end; local a, b = 10, 20; return a * b`,
	},
	{
		priority: models.HighPriority,
		script:   `local start = os.time(); while os.time() - start < 5 do end; local a, b = 10, 20; return a * b`,
	},
}
```

**Location:**
- Test data defined inline within test functions
- No separate fixtures directory or factory pattern
- Simple struct definitions used for test input data

## Coverage

**Requirements:**
- No code coverage requirement file detected
- No coverage threshold enforcement configured

**View Coverage:**
```bash
go test -cover ./...                    # Show coverage percentage
go test -coverprofile=coverage.out ./...; go tool cover -html=coverage.out  # Generate HTML report
```

## Test Types

**Unit Tests:**
- Scope: Individual functions and methods
- Approach: Direct function calls with test data
- Execution: Fast, no external dependencies
- Example: Testing Lua task execution, scheduler job queue operations

**Integration Tests:**
- Scope: Multiple components working together
- Approach: Spinning up servers, connecting to RabbitMQ
- Location: `server/tests/`, `worker/tests/`, `server/messaging/tests/`
- Execution: Slower, requires RabbitMQ running locally
- Example from `worker/tests/HealthCheck_test.go`: Full health reporting flow with RabbitMQ

**E2E Tests:**
- Framework: Not formally structured
- Pattern: Integration tests serve as E2E tests
- RabbitMQ dependency: Tests directly use RabbitMQ connection
- Example from `server/tests/api_scheduler_test.go`: Full job submission and scheduling flow

## Common Patterns

**Async Testing:**

Channel-based pattern from `server/messaging/tests/RabbitFunctions_test.go`:
```go
func Test_Rabbit(t *testing.T) {
	r, err := messaging.NewRabbit(conn)
	ch1 := make(chan Rabbit.HealthReportWrapper, 20)
	ch2 := make(chan Rabbit.RegistrationWrapper, 20)

	go r.ListenHeartBeat(ch1)
	go r.ListenRegister(ch2)

	select {
	case n = <-ch2:
		workerId = n.WorkerId
		fmt.Println("Получен workerId:", workerId)
	case <-time.After(10 * time.Second):
		fmt.Println("Failed to get Id")
		return
	}
}
```

Patterns:
- Goroutines launched with `go` keyword
- Buffered channels used: `make(chan Type, capacity)`
- Select statement for channel operations with timeout

**Error Testing:**

From `worker/executor/Executor.go`:
```go
func (e *Executor) Task(body string, workerId string) (interface{}, error) {
	l := lua.NewState()
	lua.OpenLibraries(l)

	if err := lua.DoString(l, body); err != nil {
		return nil, fmt.Errorf("lua execution error: %w", err)
	}
	// ... rest of implementation
}
```

Error testing approach:
- Wrap errors with context: `fmt.Errorf("description: %w", err)`
- Return error as last return value: `(value, error)`
- In tests, verify error presence via:
  - Direct error checks: `if err != nil { ... }`
  - Message inspection via string conversion

**Test Helpers:**

Pattern from `worker/tests/HealthCheck_test.go`:
```go
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
```

Usage: `failOnError(err, "Failed to connect to RabbitMQ")`

- Custom error handlers for critical setup failures
- Panic used to stop test immediately on failure
- Human-readable error messages with context

**Timeouts:**

Pattern from tests:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

select {
case result := <-ch:
	// Process result
case <-time.After(60 * time.Second):
	fmt.Println("Test timeout reached")
}
```

- Context timeouts for bounded operations
- Select/time.After for test execution timeouts
- Prevents tests from hanging indefinitely

## Test Execution Requirements

**External Dependencies:**
- RabbitMQ server running on `amqp://guest:guest@localhost:5672/`
- Local HTTP server (port 8080) for API tests
- No Docker setup detected; requires manual service startup

**Test Signals:**
- Tests wait for OS signals (SIGINT, SIGTERM) for graceful shutdown
- Used to manually terminate long-running tests

---

*Testing analysis: 2026-03-24*
