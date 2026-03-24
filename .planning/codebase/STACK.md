# Technology Stack

**Analysis Date:** 2026-03-24

## Languages

**Primary:**
- Go 1.24.2 - Server and worker applications

## Runtime

**Environment:**
- Alpine Linux (containerized)

**Package Manager:**
- Go modules
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Gorilla Mux 1.8.1 - HTTP routing and API handler (`server/api/handler.go`, `server/api/routes.go`)

**Messaging:**
- RabbitMQ AMQP 091 v1.10.0 - Message queue protocol for distributed task scheduling

**Scripting:**
- Shopify go-lua 0.0.0-20240527182111-9ab1540f3f5f - Lua script execution engine used in worker executor (`worker/executor/Executor.go`)

**Logging:**
- Uber Zap 1.27.0 - Structured logging (`pkg/logger/logger.go`)

**Utilities:**
- Google UUID 1.6.0 - UUID generation for worker IDs
- golang-collections 0.0.0-20130729185459-604e922904d3 - Data structure utilities

## Key Dependencies

**Critical:**
- `github.com/rabbitmq/amqp091-go` v1.10.0 - AMQP client for RabbitMQ communication. Enables async job distribution across workers
- `go.uber.org/zap` v1.27.0 - Production-grade logging. Provides leveled logging with structured fields
- `github.com/Shopify/go-lua` v0.0.0-20240527182111-9ab1540f3f5f - Lua runtime. Required for executing Lua scripts sent by scheduler
- `github.com/gorilla/mux` v1.8.1 - HTTP request routing. Handles API endpoints for job submission and status queries

**Infrastructure:**
- `github.com/google/uuid` v1.6.0 - UUID generation for unique worker identification
- `go.uber.org/multierr` v1.10.0 - Error handling utility (indirect dependency)

## Configuration

**Environment:**
- RabbitMQ host: Configurable via CLI flag `-rmq` (default: `amqp://guest:guest@localhost:5672/`)
- Logger environment: Set at server startup (`development` or `production`)
- Port: Fixed at 8080 for HTTP server

**Build:**
- Multi-stage Docker build in `Dockerfile`
  - Builder stage: Uses `golang:1.24-alpine`
  - Runtime stages: Uses `alpine:latest` for both server and worker containers
  - CGO disabled for static binaries: `CGO_ENABLED=0`

**Entry Points:**
- Server: `cmd/server/main.go` - Starts HTTP API server and scheduler
- Worker: `cmd/worker/main.go` - Starts worker node that listens for tasks via RabbitMQ

## Platform Requirements

**Development:**
- Go 1.24.2 or compatible
- RabbitMQ server (for local testing)

**Production:**
- Docker container runtime
- RabbitMQ broker (configured via environment)
- Alpine Linux base image
- No external databases required (state managed in-memory)

---

*Stack analysis: 2026-03-24*
