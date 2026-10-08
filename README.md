# Automax Registry

Registry service for Things, Channels, Channel ACL, and Groups. REG-01 stores all six tables in one Postgres database so foreign keys and multi-row writes stay transactional.

## Technology Stack

- **Language:** Go 1.25
- **Framework:** Fiber v2
- **Database:** PostgreSQL 16
- **Cache:** Redis 7
- **Messaging:** NATS JetStream
- **ORM:** GORM

## Project Structure

```
Automax-Registry/
├── cmd/server/           # Application entry point
│   └── main.go
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection and migrations
│   ├── handlers/         # HTTP request handlers
│   ├── services/         # Business logic
│   ├── repository/       # Data access layer
│   ├── models/           # Data models
│   └── natsclient/       # Shared NATS core and JetStream client
├── pkg/utils/            # Utility functions
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## NATS

Registry uses the same JetStream-enabled NATS cluster as Core and Dispatch. Set `NATS_URL` to that cluster. The Compose NATS on port `4224` is the standalone stand-in and already starts with JetStream.

Registry owns stream `REGISTRY` and subjects `registry.thing.>`. It does not subscribe to Dispatch subjects. Core keeps stream `INCIDENT_STATE`.

Queue groups and durable consumers use a service prefix so the two services never join the same group:

| Service | Prefix | Example |
| --- | --- | --- |
| Registry | `registry-` | `registry-thing-events` |
| Dispatch | `dispatch-` | `dispatch-thing-events` |

A Registry subscribe call rejects a name that starts with `dispatch-`. Core publish is fire-and-forget. JetStream publish is stored on `REGISTRY`. Thing create, update, and delete still publish through JetStream, and fall back to core NATS when JetStream is down.

## Prerequisites

- Go 1.25+
- Docker

## Environment Variables

Defaults match `docker-compose.yml`. Create a `.env` from `.env.example` only when a port or password must change:

```bash
cp .env.example .env
```

```env
SERVER_PORT=8082
SERVER_HOST=0.0.0.0

DB_HOST=localhost
DB_PORT=5440
DB_USER=registry
DB_PASSWORD=registry
DB_NAME=registry
DB_SSLMODE=disable

REDIS_HOST=localhost
REDIS_PORT=6382
REDIS_PASSWORD=
REDIS_DB=0

NATS_URL=nats://localhost:4224
```

## Installation & Running

### Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd Automax-Registry
   ```

2. **Start Postgres, Redis, NATS, and the API container**
   ```bash
   docker-compose up -d --build
   ```

   Later starts, when the image is already built, are `docker-compose up -d`. The API listens on `http://localhost:8082`. Postgres is on host port `5440`, Redis on `6382`, and NATS on `4224`.

   Check connectivity:

   ```bash
   curl -s http://localhost:8082/health
   ```

   `db`, `redis`, and `nats` are each `up` or `down`.

3. **Or run the application on the host**

   Stop the API container first so port 8082 is free, then start the dependencies and the process:

   ```bash
   docker-compose stop api
   docker-compose up -d db redis nats
   go run ./cmd/server/main.go
   ```

`go run` downloads modules on its own.

`device_credentials` and `thing_external_refs` are created by the migration. Credential routes are `POST` and `GET /registry/things/:id/credentials` and `DELETE /registry/things/:id/credentials/:credentialId`. The plaintext secret is returned only by that `POST`. External-ref routes belong to REG-13.

### Restart

Restart the API container:

```bash
docker-compose restart api
```

If the API was started with `go run`, press Ctrl+C and run `go run ./cmd/server/main.go` again.

Restart Postgres, Redis, and NATS without deleting data:

```bash
docker-compose restart
```

### Stop

Stop the API container, and leave Postgres, Redis, and NATS running:

```bash
docker-compose stop api
```

If the API was started with `go run`, press Ctrl+C in that terminal.

Stop every Compose service and keep the data:

```bash
docker-compose stop
```

### Down

Remove the containers and keep the database volume:

```bash
docker-compose down
```

Remove the containers and delete the database volume:

```bash
docker-compose down -v
```

After `stop` or `down`, start again with `docker-compose up -d`. A new volume from `down -v` is empty until the API runs its migration and creates the six tables.

## Test

```bash
go test ./...
```

API tests start a Postgres container named `automax-registry-test-pg` on `127.0.0.1:5441`, or use `REGISTRY_TEST_DSN` when that variable is set.
