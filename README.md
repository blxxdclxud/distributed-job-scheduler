# Dnp25-project-19
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## Name
Distributed Job Scheduler

## Description
The goal of the project is a development of a distributed task scheduling system designed for efficient management and distribution of computing loads in a multi-node environment. The system receives tasks from clients, distributes them between available worker nodes, and monitors the state of each node in real time.
Implemantations aims to create a fault-tolerant and scalable architecture that guarantees correct execution of tasks, even if individual system components fail.

## Execution Examples

### Basic job scheduling

Submit a Lua script and poll until the worker returns the result:

```bash
go build -o client cmd/client/main.go
./client -file lua-examples/factorial.lua -priority 0
```

```
Submitting job to server at http://localhost:8080...
Job submitted successfully! Job ID: 1, Initial status: QUEUED
Polling for job result...
Current status: QUEUED
Current status: IN_PROGRESS
Current status: COMPLETED
Job completed! Result: 7886578673647905035523632139321850622951359776871732632947425332443594499634033429203042840119846239041772121389196388302576427902426371050619266249528299311134628572707633172373969889439224456214516642402540332154
```

### Job rescheduling

If a worker goes down mid-execution the scheduler detects the missed heartbeat and reassigns the job to another available worker:

```bash
./client -file lua-examples/downtest.lua -priority 1
```

```
Submitting job to server at http://localhost:8080...
Job submitted successfully! Job ID: 2, Initial status: QUEUED
Polling for job result...
Current status: QUEUED
Current status: IN_PROGRESS
Current status: IN_PROGRESS
Current status: QUEUED        ← worker-1 went down, job rescheduled
Current status: IN_PROGRESS   ← picked up by worker-2
Current status: COMPLETED
Job completed! Result: 3
```

## Installation
```bash
git clone https://github.com/blxxdclxud/distributed-job-scheduler.git
cd distributed-job-scheduler
```

## Usage
### Host and Workers
Getting the project up using docker compose
```bash
cd deployments
docker-compose -f docker-compose.yml build
docker-compose -f docker-compose.yml up
```
### Client
```bash
go build cmd/client/main.go
./main -file lua-examples/factorial.lua
```

## Authors and acknowledgment
- **Egor Pustovoytenko** e.pustovoytenko@innopolis.university (Report, presentation and initial design)

- **Askar Dinikeev** a.dinikeev@innopolis.university (Initial and final design, implementation, demo)

- **Niyaz Gubaidullin** n.gubaidullin@innopolis.university (Implemantation)

- **Ramazan Nazmiev** r.nazmiev@innopolis.university (Implemantation)

- **Nurzhan Baxikov** n.baxikov@innopolis.university (Management, documentation)

## License
[MIT LICENSE](https://gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/-/blob/main/LICENSE)

