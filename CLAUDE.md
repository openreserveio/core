# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Finsorb is a collection of financial microservices (Open Source) for building headless ledgers, general ledgers, and payment systems. Each service is an independent Go module with its own `go.mod`, living under `src/`.

## Language & Toolchain

- **Go 1.26.x** — each module declares its own version in `go.mod`
- **Protocol Buffers / gRPC** — inter-service communication; source `.proto` files in each module's `model/` or `glmodel/` directory
- **NATS JetStream** — async messaging bus used by `core-gl`, `core-ledger-poster`, and `core-payments`
- **PostgreSQL via uptrace/bun** — ORM used across all data-bearing services
- **Cobra** — CLI framework for all service entrypoints

## Architecture

Services are layered; higher layers call lower layers via gRPC or NATS:

```
core-external-api-service  (REST/gRPC gateway for external clients)
        ↓ gRPC
core-gl          core-payments
        ↓ NATS bus (ledger-queue)
core-ledger-poster          ← consumes NATS, posts to core-ledger via gRPC
        ↓ gRPC
core-ledger                 ← only service with direct DB writes
        ↓
    PostgreSQL (LedgerDB)
```

- **`core-ledger`** — foundational service; owns `Ledger`, `LedgerTransaction`, `Account`, `AccountBalance` domain objects and all PostgreSQL writes. Exposes a gRPC server on port `4080` by default.
- **`core-gl`** — general ledger layer; manages `ChartOfAccounts`, `Entity`, and GL-level operations (journal entries, FedNow payments). Calls `core-ledger` via gRPC and publishes/subscribes to NATS. Has its own entity Postgres DB.
- **`core-ledger-poster`** — NATS microservice (registered via `nats.go/micro`) that bridges async NATS messages onto synchronous `core-ledger` gRPC calls.
- **`core-payments`** — payment rail framework with sub-commands per rail (`fednow-connector`, `fednow-inbound-payment-workflow`, etc.). Uses both NATS and gRPC.
- **`core-external-api-service`** — HTTP REST + gRPC gateway for external clients; uses OpenAPI spec at `extapimodel/external-api.oas.yaml`.
- **`core-fintech-programs`** — FinTech program management service (standalone or integrated).
- **`core-util`** - Shared Logic between all core services like logging, OTEL, etc
- **`integration-tests`** — standalone Go module with Ginkgo/Gomega tests that connect to live running services.

### Internal package conventions

Every service follows the same layout:
- `model/`, `glmodel/`, `ftpmodel/`, `pmtmodel/` — `.proto` source files and hand-written domain types
- `generated/` — compiled gRPC/protobuf Go code (never edit manually)
- `service/` — business logic; one file per operation (e.g. `post_ledger_transaction.go`)
- `bus/` — NATS connect/send/receive helpers (copy-consistent across services)
- `application/<servicename>/cmd/` — Cobra commands; `start.go` wires flags → service constructor → gRPC server

## Common Commands

Run these from within the relevant service directory (e.g. `cd core-ledger`).

### Build

```sh
go build ./...
```

### Generate gRPC code from .proto files

```sh
make generate-grpc
```

Requires `protoc` and `protoc-gen-go` / `protoc-gen-go-grpc` installed.

### Build Docker image

```sh
make dockerize
```

### Run tests (integration)

Integration tests live in `integration-tests/` and require running infrastructure. From `integration-tests/`:

```sh
# Run all suites
go test ./...

# Run a single suite (e.g. core-gl)
go test ./core-gl/...

# Run with Ginkgo directly (verbose output)
ginkgo -v ./core-gl/...
```

### Run a service locally

**core-ledger** (gRPC on :4080):
```sh
go run ./application/coreledger/main.go start \
  --dburl "postgresql://finsorbuser:finsorbpass@localhost:5432/coreledgerdb?sslmode=disable" \
  --listenHost 0.0.0.0 --listenPort 4080
```

**core-gl** (gRPC on :4081, connects to NATS node 2 and core-ledger):
```sh
go run ./application/coregl/main.go start \
  --entitydburl "postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable" \
  --busconnurl nats://localhost:4322 \
  --coreledgerurl localhost:4080 \
  --listenHost 0.0.0.0 --listenPort 4081
```

**core-ledger-poster** (NATS microservice, connects to NATS node 3):
```sh
go run ./application/coreledgerposter/main.go start \
  --natsurl nats://localhost:4422 \
  --coreledgerurl 0.0.0.0:4080
```

**core-payments fednow-connector**:
```sh
go run ./application/corepayments/main.go fednow-connector \
  --paymentsdburl "postgresql://finsorbuser:finsorbpass@localhost:5432/paymentsdb?sslmode=disable" \
  --busconnurl nats://localhost:4422 \
  --listenHost 0.0.0.0 --listenPort 8083
```

## Infrastructure

Spin up the full local infrastructure (Postgres, Redis, NATS cluster) with:

```sh
cd infrastructure/docker
docker compose -f infrastructure-compose.yaml up -d
```

Default credentials and ports:
- **PostgreSQL**: `localhost:5432`, user `finsorbuser`, pass `finsorbpass`, db `coreledgerdb`
- **NATS**: cluster of 3 nodes — `nats://localhost:4222`, `4322`, `4422`
- **Redis**: `localhost:6379`
- **OTEL Collector**: `localhost:4318` for OTLP http, `localhost:4317` for OTLP grpc

## Code Generation

After modifying any `.proto` file, regenerate the Go code with `make generate-grpc` from that service's directory. The `generated/` directory is committed — update it whenever protos change.
