# Relay

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-336791?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Apache Kafka](https://img.shields.io/badge/Kafka-3.0+-231F20?style=flat&logo=apachekafka&logoColor=white)](https://kafka.apache.org/)
[![Redis](https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A distributed, fault-tolerant job execution system built in Go. Relay handles job scheduling, execution, and orchestration with support for retries, job chaining, and distributed locking.

## Overview

Relay is designed to reliably execute background jobs across distributed workers. It combines PostgreSQL for durable storage, Kafka for distributed messaging, and Redis for distributed locking to ensure jobs are executed exactly once, even in the presence of failures.

### Key Features

- **Reliable Job Execution** — Jobs are persisted before processing, ensuring no work is lost
- **Distributed Processing** — Scale horizontally with multiple workers consuming from Kafka
- **Exactly-Once Semantics** — Redis-based distributed locking prevents duplicate execution
- **Shell Task Execution** — Run external scripts and binaries with stdout/stderr capture
- **Automatic Retries** — Configurable retry logic with exponential backoff
- **Job Chaining** — Define workflows where completing one job triggers the next
- **Dead Letter Queue** — Failed jobs are quarantined for manual inspection
- **CLI Interface** — Submit, monitor, and manage jobs from the command line

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Client     │────▶│   Relay API  │────▶│  PostgreSQL  │
│   (CLI)      │     │              │     │  (Storage)   │
└──────────────┘     └──────┬───────┘     └──────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │    Kafka     │
                    │   (Queue)    │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
       ┌──────────┐ ┌──────────┐ ┌──────────┐
       │ Worker 1 │ │ Worker 2 │ │ Worker N │
       └────┬─────┘ └────┬─────┘ └────┬─────┘
            │            │            │
            └────────────┼────────────┘
                         ▼
                  ┌──────────────┐
                  │    Redis     │
                  │   (Locks)    │
                  └──────────────┘
```

## Job Lifecycle

```
PENDING ──▶ RUNNING ──▶ COMPLETED
                │
                ├──▶ FAILED ──▶ (retry) ──▶ PENDING
                │
                └──▶ DEAD (after max retries)
```

1. **PENDING** — Job is created and queued for execution
2. **RUNNING** — Worker has acquired the lock and is executing the job
3. **COMPLETED** — Job finished successfully
4. **FAILED** — Job execution failed, may be retried
5. **DEAD** — Job exhausted all retries, moved to dead letter queue

## Installation

### Prerequisites

- Go 1.21+
- PostgreSQL 16+
- Apache Kafka 3.0+
- Redis 7.0+

### Build from Source

```bash
git clone https://github.com/tomiwa-a/Relay.git
cd relay
go build -o relay ./cmd/api
```

### Database Setup

```bash
# Run migrations
make db/migrations/up
```

## Configuration

Relay is configured via environment variables or command-line flags:

| Variable              | Flag             | Default          | Description                  |
| --------------------- | ---------------- | ---------------- | ---------------------------- |
| `RELAY_DB_DSN`        | `-db-dsn`        | —                | PostgreSQL connection string |
| `RELAY_KAFKA_BROKERS` | `-kafka-brokers` | `localhost:9092` | Kafka broker addresses       |
| `RELAY_REDIS_ADDR`    | `-redis-addr`    | `localhost:6379` | Redis server address         |
| `RELAY_PORT`          | `-port`          | `4000`           | API server port              |
| `RELAY_ENV`           | `-env`           | `development`    | Environment mode             |

## Usage

### Starting the Server

```bash
make run/api
```

### CLI Commands

```bash
# Submit a job from a JSON file
relay submit job.json

# List jobs with optional status filter
relay list --status=pending
relay list --status=failed

# View logs for a specific job
relay logs <job_id>

# Retry a failed job
relay retry <job_id>
```

### Job Payload Example

```json
{
  "type": "SHELL",
  "payload": {
    "command": "/usr/local/bin/process-data.sh",
    "args": ["--input", "/data/file.csv"],
    "timeout": "5m"
  },
  "on_success": {
    "type": "SHELL",
    "payload": {
      "command": "/usr/local/bin/notify.sh"
    }
  },
  "max_retries": 3,
  "backoff": "exponential"
}
```

## Development

### Project Structure

```
relay/
├── cmd/
│   └── api/           # Application entrypoint
├── internal/
│   ├── api/           # HTTP handlers and routing
│   ├── worker/        # Kafka consumer and job execution
│   ├── executor/      # Shell and task executors
│   ├── repository/    # Database access (sqlc generated)
│   └── queries/       # SQL query definitions
├── migrations/        # Database migrations
└── Makefile
```

## Roadmap

- [x] Phase 1: Job ingestion, persistence, and basic execution
- [ ] Phase 2: Kafka distribution and Redis locking
- [ ] Phase 3: Shell executor with output capture
- [ ] Phase 4: Retry logic, job chaining, and dead letter queue
- [ ] Phase 5: CLI interface

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
