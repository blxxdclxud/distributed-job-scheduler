# Codebase Concerns

**Analysis Date:** 2026-03-24

## Tech Debt

**Incomplete GetJob Implementation:**
- Issue: `GetJob()` method in scheduler is a stub that returns job directly from map without proper error handling or state validation
- Files: `server/scheduler/scheduler.go:132-135`
- Impact: API endpoint `/jobs/{id}` cannot properly report missing jobs or handle state errors, returns nil error even when job doesn't exist
- Fix approach: Implement proper validation to check if job exists before returning, return appropriate error when not found, check if job status/state is valid

**Inconsistent Logging Patterns:**
- Issue: Mixture of `fmt.Println`, `fmt.Printf`, `logger.Info`, and `zap.Error` throughout codebase instead of unified logging approach
- Files: `server/messaging/rabbit.go:65,70,75`, `server/messaging/ListenTaskResults.go:20,33`, `server/messaging/ListenRegister.go:21,31`, `server/messaging/ListenHeartBeat.go:33`, `cmd/server/main.go:18`, `cmd/worker/main.go:39`, multiple test files
- Impact: Inconsistent log levels, lost structured logging in some paths, difficult debugging and monitoring, logs go to stdout instead of configured handlers
- Fix approach: Replace all `fmt.Print*` calls with logger methods, standardize on single logging framework (zap via pkg/logger), ensure all logging passes through configured logger

**Incomplete Server Initialization:**
- Issue: `RunServer()` in `server/server.go` is marked as "just placeholder now !!!" - scheduler and RabbitMQ client are not wired together
- Files: `server/server.go:11-22`
- Impact: Scheduler created but never connected to RabbitMQ client. `SetRabbitClient()` is never called, meaning tasks cannot be sent to workers
- Fix approach: Complete server initialization to create Rabbit client, pass it to scheduler, initialize all message listeners (heartbeat, results, registration)

**Unused Logger Initialization Parameter:**
- Issue: `cmd/server/main.go` creates logger but doesn't use it
- Files: `cmd/server/main.go:20-22`
- Impact: Logger is created but server uses hardcoded panic calls instead of proper logging
- Fix approach: Pass logger to RunServer(), use logger for all error handling

---

## Known Bugs

**Error Ignored in Worker Executor Initialization:**
- Symptoms: First error assignment in `NewExecutor()` is silently overwritten
- Files: `worker/executor/Executor.go:24-31`
- Trigger: When publishing to either ResultExchange or WorkerStatusExchangeName fails, first error is lost and only second error matters
- Code flow: `p, err :=` (line 25) → `ack, err :=` (line 26) overwrites first error before checking
- Impact: Publisher creation failures partially hidden, only last error visible when both fail
- Fix: Use separate error variables or check err after each call: `if err != nil { panic(err) }`

**Unused Variable in Worker Main:**
- Symptoms: Error variable declared but not checked before use
- Files: `cmd/worker/main.go:45-49`
- Trigger: `if err != nil` blocks with no action, errors silently continue to next check
- Impact: If registration publisher fails to create (line 55), code continues with nil publisher and will panic at line 58
- Fix approach: Check and handle error immediately after `NewRabbitMQPublisher` call on line 55

**RabbitMQ Channel Not Closed:**
- Symptoms: Rabbit channel opened in multiple places but never explicitly closed
- Files: `server/messaging/rabbit.go:19` opens channel but no defer close, `worker/executor/Executor.go:35` opens channel but no close
- Impact: Resource leak under normal operation, channels accumulate until connection exhaustion or reconnection issues
- Fix: Add `defer ch.Close()` after successful channel creation in `ListenTasks()` and `NewRabbit()`

**Missing Nil Check in Worker Assignment:**
- Symptoms: Code accesses worker without checking if `RoundRobin()` actually returns valid worker
- Files: `server/scheduler/scheduler.go:78-86, 93-98`
- Trigger: When `RoundRobin()` returns nil (no available workers), code proceeds to dereference
- Impact: If task assignment called when workers list is empty, `worker.ID` access will panic
- Fix: Add explicit check: `if worker == nil { return }` in both `AssignTask()` and `ReassignTask()`

**Message Channel Leaks in Listeners:**
- Symptoms: Listener goroutines receive messages on infinite loops with no way to stop
- Files: `server/messaging/ListenHeartBeat.go:28-43`, `server/messaging/ListenTaskResults.go:28-42`, `server/messaging/ListenRegister.go:26-36`
- Trigger: Server shutdown or connection close
- Impact: Goroutines continue running after channel is closed, may cause panics on send to closed channel or consume from closed queue
- Fix: Add stop channel as parameter, check for stop signal in select statement

---

## Security Considerations

**Arbitrary Lua Code Execution:**
- Risk: Worker directly executes Lua code from messages without sandboxing or restrictions
- Files: `worker/executor/Executor.go:122-134`
- Current mitigation: None - any Lua code received is executed
- Attack scenario: Malicious actor sends Lua code that accesses system resources, filesystem, or performs denial of service
- Recommendations:
  - Implement Lua sandbox with restricted libraries (no OS, IO, debug modules)
  - Add code length limits to prevent resource exhaustion
  - Implement timeout on Lua script execution
  - Consider using a safer scripting language or pre-compiled bytecode verification

**Hardcoded RabbitMQ Credentials:**
- Risk: Default RabbitMQ credentials hardcoded in command-line flags
- Files: `cmd/server/main.go:15`, `cmd/worker/main.go:36`
- Current mitigation: Can be overridden via `-rmq` flag but defaults to guest:guest
- Recommendations:
  - Remove hardcoded credentials from source
  - Require env vars for RabbitMQ connection (e.g., `RABBITMQ_URL`)
  - Use TLS for RabbitMQ connections in production
  - Document credential management in deployment guide

**No Input Validation on Job Scripts:**
- Risk: Job submission handler accepts script without validation
- Files: `server/api/handler.go:22-40`
- Current mitigation: None
- Recommendations:
  - Add script size limits (max MB)
  - Validate script format before queuing
  - Implement rate limiting on job submissions per client
  - Add authentication to API endpoints

**No Authentication on API Endpoints:**
- Risk: Any client can submit and query jobs
- Files: `server/api/routes.go` (routes registered without auth middleware)
- Current mitigation: None
- Recommendations:
  - Implement API key or token-based authentication
  - Add role-based access control if multiple user types needed
  - Implement rate limiting per client

---

## Performance Bottlenecks

**Synchronous Job Status Updates:**
- Problem: Job status updates in scheduler require full mutex lock on entire scheduler state
- Files: `server/scheduler/scheduler.go:47-70`
- Cause: Single mutex protects both workers and jobs, status update blocks all scheduling operations
- Observed behavior: When worker fails and task is reassigned, entire scheduler locked
- Improvement path: Use separate mutexes for workers and jobs, or implement lock-free queue for status updates

**No Worker Health Check Recovery:**
- Problem: Dead workers remain in rotation until next assignment attempt fails
- Files: `server/scheduler/scheduler.go`, no auto-removal of failed workers
- Current behavior: Failed worker assigned task, assignment fails, task re-queued, but worker stays in queue
- Impact: At scale, dead workers create cascading failures and task delays
- Improvement path: Integrate heartbeat listener with scheduler to mark workers unhealthy, implement exponential backoff for repeated failures

**Task Result Processing Unbuffered:**
- Problem: TaskReplyWrapper channel in listeners has no buffer and sender doesn't handle blocking
- Files: `server/messaging/ListenTaskResults.go:9-43`
- Cause: Channel send `c <- message` can block if receiver is slow
- Impact: Result processing delays cascade back to worker, blocking task completion acknowledgment
- Improvement path: Use buffered channel with size based on expected message rate, implement overflow handling

**Unbounded In-Memory Job Storage:**
- Problem: All jobs stored in `AllJobs` map forever without cleanup
- Files: `server/scheduler/scheduler.go:24`, line 36 initializes map that never removes completed jobs
- Cause: No job cleanup or archival mechanism
- Impact: At scale with high job volume (thousands/day), memory usage grows unbounded and causes OOM
- Improvement path: Implement job archival after completion (e.g., after 24 hours), store only active jobs in memory, use persistent storage for history

---

## Fragile Areas

**RabbitMQ Reconnection Logic Missing:**
- Files: `server/messaging/rabbit.go`, `worker/executor/Executor.go`, `worker/messaging/Rabbit.go`
- Why fragile: Single connection used throughout - if RabbitMQ restarts or network drops, entire system stops
- Current behavior: Connection errors logged but no retry or reconnection attempt
- Safe modification: Implement connection pooling with automatic reconnect, add circuit breaker for failed connections
- Test coverage: No tests for connection failures or reconnection scenarios

**Scheduler State Inconsistency:**
- Files: `server/scheduler/scheduler.go`
- Why fragile: Multiple independent structures (AvailableWorkers, TotalWorkers, AllJobs, WorkerAssignments) must stay in sync but no invariant checking
- Unsafe modification: Adding/removing workers without updating all data structures simultaneously can cause inconsistency
- Example: Worker removed from AvailableWorkers but still in TotalWorkers causes assignment to ghost worker
- Safe modification: Create invariant checks on scheduler state, use transaction-like pattern for multi-structure updates
- Test coverage: No tests for concurrent scheduler modifications

**Message Ordering Assumptions:**
- Files: `server/messaging/ListenTaskResults.go`, `server/messaging/ListenHeartBeat.go`
- Why fragile: Code assumes messages arrive in order (task sent, then result received) but RabbitMQ topic exchange doesn't guarantee order
- Risk: Results processed before tasks are properly tracked, or heartbeats arrive out of sync
- Safe modification: Add sequence numbers or timestamps to messages, implement out-of-order handling
- Test coverage: No tests for out-of-order message processing

**Lua Execution Environment Not Isolated:**
- Files: `worker/executor/Executor.go:122-134`
- Why fragile: Each task creates new Lua state but doesn't clean up properly - potential for resource leaks across tasks
- Risk: Long-running worker accumulates Lua states, memory grows unbounded
- Safe modification: Implement proper Lua state cleanup, add memory/execution limits per task
- Test coverage: No stress tests for repeated Lua execution

---

## Scaling Limits

**Single Scheduler Instance:**
- Current capacity: Handles limited number of concurrent tasks and workers
- Limit: Scheduler is single point of failure and bottleneck - cannot be horizontally scaled
- Architecture issue: All job assignment logic centralized in one `Scheduler` instance
- Scaling path: Implement distributed scheduler or use message-based job assignment (workers pull jobs from queue instead of scheduler pushing)

**Worker Round-Robin Not Distributed:**
- Current approach: Server-side round-robin scheduling
- Limit: Scheduler must track all worker states, doesn't scale beyond hundreds of workers
- Scaling path: Implement pull-based model where workers consume from shared job queue, eliminating need for server-side state

**Message Queue Not Tuned:**
- Files: RabbitMQ exchange/queue declarations in `server/messaging/rabbit.go`, `server/utils/InitRabbit/*`
- Current config: All queues auto-created but no throughput tuning
- Limit: Default queue settings may not handle thousands of jobs/sec
- Scaling path: Implement queue prefetch limits, set appropriate acknowledgment modes, consider queue sharding by priority

---

## Dependencies at Risk

**Deprecated golang-collections Library:**
- Risk: `github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3` is unmaintained (last update 2013)
- Impact: No security updates, thread-safety issues if used in concurrent context
- Current usage: Queue implementation in `server/scheduler/task_queue.go`, `server/scheduler/worker_queue.go`
- Migration plan: Replace with standard Go channels or `sync.Map`, or use well-maintained queue library like `github.com/coocood/freecache`

**go-lua Library Potential Issues:**
- Risk: `github.com/Shopify/go-lua` is Shopify-specific, may have edge cases for non-Shopify usage
- Impact: Lua execution errors not handled, potential sandbox escapes
- Current usage: `worker/executor/Executor.go:123-124`
- Migration plan: Consider using `github.com/yuin/gopher-lua` (more actively maintained) or implement strict code review for security

---

## Missing Critical Features

**Job Failure Handling:**
- Problem: No mechanism for failed jobs to be retried or reported to client
- Blocks: Cannot guarantee job execution reliability, client has no way to know if job ultimately succeeded
- Current state: Task reassignment exists but no max retry limit or failure notification
- Implementation needed: Add retry counter, implement exponential backoff, send failure notification to API client

**Worker Health Monitoring:**
- Problem: Heartbeat received but not used to detect dead workers
- Blocks: Dead workers cause cascading failures, no proactive worker removal
- Current state: Heartbeats listened but `ListenHeartBeat` never integrated with scheduler
- Implementation needed: Track heartbeat timestamps, implement timeout detection, auto-remove workers with stale heartbeats

**Job Priority Actually Not Used in Assignment:**
- Problem: While jobs stored in priority queues, round-robin worker selection doesn't consider priority
- Blocks: High-priority jobs may wait behind low-priority if same worker assigned
- Current state: Jobs prioritized in queue but once assigned to worker, execution order depends on worker's queue
- Implementation needed: Implement priority scheduling at worker level or centralized scheduling aware of priorities

**Result Storage and Retrieval:**
- Problem: Task results processed and printed but not stored for client retrieval
- Blocks: Client cannot retrieve job results after completion
- Current state: `ListenTaskResults.go` receives results but doesn't persist them
- Implementation needed: Store results in persistent storage, implement job results API endpoint

**API Request Validation:**
- Problem: No validation of input parameters
- Blocks: Malformed requests could crash handlers
- Current state: Direct type assertion without bounds checking
- Implementation needed: Add input validation, bounds checking, type validation middleware

---

## Test Coverage Gaps

**No Tests for Scheduler State Management:**
- What's not tested: Worker assignment, task queuing under concurrent access
- Files: `server/scheduler/scheduler.go`
- Risk: Concurrent modifications could cause race conditions, data loss, or inconsistent state
- Priority: High - core scheduling logic untested

**No Tests for Message Processing:**
- What's not tested: Correct handling of heartbeats, task results, worker registration
- Files: `server/messaging/ListenHeartBeat.go`, `server/messaging/ListenTaskResults.go`, `server/messaging/ListenRegister.go`
- Risk: Silent failures, corrupted state, message order issues undetected
- Priority: High - message handling is critical path

**No Integration Tests:**
- What's not tested: Full job lifecycle from submission to result delivery
- Current tests: Only unit test in `server/tests/api_scheduler_test.go` is incomplete (creates server but doesn't test response)
- Risk: System may work in isolation but fail when integrated
- Priority: High - integration is where most real bugs occur

**No Error Scenario Tests:**
- What's not tested: Network failures, RabbitMQ disconnection, worker crashes, malformed messages
- Risk: Error handling code paths never executed, production failures unpredictable
- Priority: Medium - error paths critical but untested

**No Load/Stress Tests:**
- What's not tested: Performance with thousands of concurrent jobs, message backlogs, worker scaling
- Risk: Performance bottlenecks and scaling limits unknown until production
- Priority: Medium - scaling issues discovered too late

**No Lua Script Security Tests:**
- What's not tested: Sandboxing of Lua execution, resource limits, malicious code handling
- Files: `worker/executor/Executor.go:122-134`
- Risk: Security vulnerabilities in Lua execution undetected
- Priority: High - security critical

---

*Concerns audit: 2026-03-24*
