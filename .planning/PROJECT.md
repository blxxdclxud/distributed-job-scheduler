# Distributed Job Scheduling — Portfolio Metrics

## What This Is

Go-based distributed job scheduling system, где сервер принимает задачи через HTTP API, распределяет их по воркерам через RabbitMQ, используя приоритетные очереди и round-robin балансировку. Воркеры исполняют Lua-скрипты и возвращают результаты. Цель проекта — получить верифицируемые метрики для резюме, а не production-деплой.

## Core Value

Реальные числа в bullet points резюме: throughput, latency, распределение нагрузки — всё подтверждённое `go test -bench` и `go test -race`.

## Requirements

### Validated

- ✓ Приоритетные очереди задач (High/Mid/Low) с гарантированным порядком обслуживания — existing
- ✓ Round-robin балансировка через WorkerQueue (FIFO-очередь воркеров) — existing
- ✓ Thread-safe доступ через sync.Mutex в EnqueueJob / AssignTask — existing
- ✓ Failover через ReassignTask при падении воркера — existing
- ✓ Асинхронная коммуникация через RabbitMQ (AMQP 0-9-1) — existing
- ✓ HTTP API для приёма задач (Gorilla Mux) — existing
- ✓ Регистрация воркеров и heartbeat мониторинг — existing
- ✓ Исполнение Lua-скриптов на воркерах — existing

### Active

- [ ] Benchmark тесты: throughput EnqueueJob и AssignTask (jobs/sec)
- [ ] Race detector тест: конкурентный EnqueueJob из N горутин под `go test -race`
- [ ] Тест корректности приоритетов: mix High/Low задач → порядок обслуживания
- [ ] Тест равномерности round-robin: N задач на M воркеров → дисбаланс ≤ X%
- [ ] Исправить nil-panic в AssignTask (CONCERNS.md: Missing Nil Check)
- [ ] Зафиксировать результаты и сформулировать 3-4 bullet points

### Out of Scope

- Production deployment — цель аналитическая, не операционная
- Frontend / dashboard — не нужен для метрик
- Персистентность состояния (БД) — in-memory достаточно для бенчмарков
- Нагрузочный тест с реальным RabbitMQ — только если есть время; unit-бенчмарки достаточны

## Context

- Командный проект (university, dnp25-project-19), автор отвечал за `server/scheduler/` — все 3 файла
- Git branch: `r.nazmiev-scheduler-logic`
- Codebase map: `.planning/codebase/` (создан 2026-03-24)
- Критичный баг для исправления перед бенчмарками: nil-dereference в `scheduler.go:78` при пустой WorkerQueue

## Constraints

- **Scope**: Только код планировщика (`server/scheduler/`) — не трогаем messaging и worker
- **Tech stack**: Стандартный `testing` пакет Go, никаких новых зависимостей
- **Time**: Бенчмарки должны быть воспроизводимы за 1 команду: `go test -bench=. ./server/scheduler/...`

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Бенчмарки только для scheduler, без RabbitMQ | Изолируем вклад автора, результат воспроизводим без инфраструктуры | — Pending |
| Исправить nil-check перед бенчмарками | Без этого BenchmarkAssignTask с 0 воркерами упадёт с panic | — Pending |
| go test -race как отдельный bullet point | Race safety — самостоятельная ценность, рекрутеры это понимают | — Pending |

---
*Last updated: 2026-03-24 after initialization*
