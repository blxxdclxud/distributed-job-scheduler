# Dnp25-project-19
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## Name
Distributed Job Scheduler

## Description
The project is the development of a distributed task scheduling system designed for efficient management and distribution of computing loads in a multi-node environment. The system receives tasks from clients, distributes them between available worker nodes, and monitors the state of each node in real time.
The goal of the project is to create a fault-tolerant and scalable architecture that guarantees correct execution of tasks, even if individual system components fail.

## Visuals

### Basic job scheduling

![](assets/dnp_demo1.mp4)

### Job rescheduling

![](assets/dnp_demo2.mp4)

## Installation
```bash
git clone https://gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19
cd Dnp25-project-19/deployments
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
- **Egor Pustovoytenko** e.pustovoytenko@innopolis.university

- **Askar Dinikeev** a.dinikeev@innopolis.university

- **Niyaz Gubaidullin** n.gubaidullin@innopolis.university

- **Ramazan Nazmiev** r.nazmiev@innopolis.university

- **Nurzhan Baxikov** n.baxikov@innopolis.university

## License
[MIT LICENSE](https://gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/-/blob/main/LICENSE)

