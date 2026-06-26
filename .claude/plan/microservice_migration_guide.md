# Microservice Migration Guide

## Migration Execution Rules

These rules are mandatory while following the migration plans.

- Do not run project commands unless the user explicitly asks for them.
- Do not run `go mod tidy`, API generation, dependency install/update, tests,
  builds, Docker, or deployment commands during migration.
- If a command would normally be useful for verification, mention it as a
  manual follow-up instead of running it.
- Do not generate unit tests or unit-test helper code unless the user explicitly
  asks for them.
- Migrate only the main source code, configuration, API contracts, and required
  wiring described by the plan.
- Keep the existing monolith code frozen. New microservice code belongs under
  the new `microservices/` workspace.

## Shared Infrastructure Design Patterns

When building shared wrapper libraries in `microservices/pkg/` (like Database or Kafka connections), strictly follow the **Functional Options Pattern** (e.g., `opts ...WriterOption`). 
- This ensures constructors remain backward-compatible when new configurations are added.
- It prevents breaking changes across multiple services when one service requires a new specific configuration.
- Shared initialization logic should be kept fully decoupled from domain logic.

## Authentication & Authorization Design

Authentication and Authorization inside the microservices workspace are decoupled to ensure loose coupling and security:

- **Authentication (Gateway Level):** The API Gateway (e.g., Traefik) intercepts external client requests, validates the client JWT, extracts claims, and injects them as HTTP headers (`X-User-Id`, `X-User-Role`) into downstream requests.
- **Authorization (Service Level - REST APIs):** Downstream services (`server-service`, `monitor-service`, `auth-service`) expose REST APIs for client-facing endpoints. They utilize HTTP Middlewares to check if the caller's role (read from `X-User-Role` header) possesses the required scopes for the requested endpoint.
- **Agent Ingestion Authentication (heartbeat-gateway):** Heartbeats sent from the remote host `agent` are authenticated at `heartbeat-gateway` using a shared host-level **API Key** (`X-API-Key` HTTP header) rather than user-level JWTs or roles, as the agent is a headless background system service.
- **Internal Service-to-Service Authorization (gRPC Streaming):** Internal system calls (e.g., `mail-worker` downloading reports from `monitor-service` via `DownloadReport` stream) still use gRPC. They bypass user JWT validation and are authenticated using a shared **internal API Key** (`x-api-key`) verified by `StreamAPIKeyInterceptor`.
- **Security Boundaries:** Zero-trust is maintained through network isolation (Swarm private overlay networks). Downstream services are not exposed to the public host ports, ensuring only authenticated traffic routed through the API Gateway or internal workers can reach them.

## Business Capabilities

The system is split around six business capabilities:

1. **Account & Access Management:** login, authorization, and JWT issuance.
2. **Server Inventory Cataloging:** management of static server records.
3. **Heartbeat Ingestion:** high-frequency heartbeat intake.
4. **Active Uptime Probing:** ICMP ping execution.
5. **Monitoring & Analytics:** collected-data processing, live state updates,
   history logging, and uptime calculation.
6. **Notification:** report formatting and email delivery.

## Bounded Contexts

The target architecture uses six bounded contexts:

| Context | Core Domain Terms |
| --- | --- |
| Auth Context | `Account`, `Token`, `Role` |
| Server Context | `ServerProfile`, `IPAddress`, `UpdatedAt` |
| Gateway Context | `RawHeartbeat` |
| Ping Context | `PingTask`, `ICMPResult` |
| Monitor Context | `LiveStatus`, `StatusLog`, `UptimeAgg`, `ReportMetrics` |
| Mail Context | `EmailPayload`, `Attachment` |

### Core Domain Term Definitions

- **`ServerProfile` (Server Context)**: Represents static server registry details (`ServerID`, `ServerName`, `IPv4`), `CreatedAt`, and `UpdatedAt` (representing the timestamp of registry changes). It explicitly excludes active monitoring statuses and history (`Status`, `LastPingAt`) to respect domain separation.
- **`LiveStatus` (Monitor Context)**: Tracks operational/uptime metrics. It records `Status` (ONLINE/OFFLINE/UNKNOWN), `LastPingAt`, and `LastHeartbeatAt`, while synchronizing name/IP alterations from server lifecycle events.
- **`StatusLog` (Monitor Context)**: Represents a single historical time-series log entry stored in Elasticsearch. It contains the server's availability state (`ServerID`, `Status`, `Timestamp`) recorded from heartbeats or active checks, used for historical uptime calculations.
- **`UpdatedAt` (Server Context)**: Denotes the timestamp when the server registry profile was last modified. This replaces the legacy `MetadataUpdatedAt` term.

## Context Mapping

Service interaction is primarily event-driven through Kafka, with an HTTP pull
pattern for report files.

- **Server Context -> Monitor Context:** `server-service` publishes
  `ServerCreated`, `ServerUpdated`, and `ServerDeleted` events. The
  `monitor-service` consumes these events to maintain its local server-status
  table and know which servers should be monitored.
- **Gateway Context -> Monitor Context:** `heartbeat-gateway` publishes
  `HeartbeatReceived` events.
- **Monitor Context -> Ping Context:** `monitor-service` checks its local cache
  for missing heartbeat signals and publishes `PingRequested`.
- **Ping Context -> Monitor Context:** `ping-worker` performs ICMP ping work and
  publishes `PingResulted`.
- **Monitor Context -> Mail Context:** `monitor-service` generates scheduled
  reports and publishes `ReportAvailable` or `ReportGenerated`.
- **Mail Context -> Monitor Context:** `mail-worker` consumes the report event,
  then actively calls the `monitor-service` internal gRPC service (`InternalFileTransferService`) to stream/download the report file and
  send it by SMTP.

## Target Microservices

The target deployment contains six microservices intended for Docker Swarm.

### 1. `auth-service`

Responsibilities:

- Manage accounts.
- Verify login.
- Issue JWTs.
- Block tokens.

Storage:

- Postgres for accounts.
- Redis for token blocklist.

### 2. `server-service`

Responsibilities:

- Provide server CRUD APIs.
- Import and export static server information.
- Publish server lifecycle events for downstream contexts.

Storage:

- Postgres for server records.

### 3. `heartbeat-gateway`

Responsibilities:

- Expose a focused `POST /heartbeat` HTTP API directly to remote agents (bypassing the main API Gateway to handle high-frequency throughput).
- Authenticate incoming agent requests directly via HTTP header (`X-API-Key`).
- Push heartbeat messages into Kafka quickly.
- Avoid database ownership.

Storage:

- No database.

### 4. `ping-worker`

Responsibilities:

- Consume ping commands from Kafka.
- Execute raw-socket ICMP ping.
- Publish ping results back to Kafka.
- Avoid database ownership.

Storage:

- No database.

### 5. `monitor-service`

Responsibilities:

- **Writer:** consume heartbeat and ping-result events from Kafka, then batch
  write current state to Postgres. It also consumes server lifecycle events from
  `server-service` to keep a local server table synchronized.
- **Logger:** consume the same operational events and batch write historical
  logs to Elasticsearch.
- **Analyzer & Checker:** maintain an in-memory cache to request ping checks
  when heartbeat data expires.
- **Reporter:** run scheduled aggregation jobs over Elasticsearch data, generate
  reports, cache report artifacts, and expose a download gRPC API (`InternalFileTransferService`) for `mail-worker`.

Storage:

- Postgres for local server information and current status.
- Elasticsearch for historical logs and report aggregation data.
- Redis for report cache.

### 6. `mail-worker`

Responsibilities:

- Consume mail/report commands from Kafka.
- Call the `monitor-service` internal gRPC service (`InternalFileTransferService`) to stream/download report files.
- Send email through SMTP.

Storage:

- No dedicated database is defined by the plan.

## Host-Level Daemons

### 1. `agent`

Responsibilities:

- Act as a lightweight daemon running on the remote monitored host servers.
- Send active heartbeat notifications at configurable intervals to `heartbeat-gateway`.

Storage:

- No database.

## Monorepo Organization

All new microservice code belongs in a new `microservices/` root so that the old
monolith remains isolated and frozen.

```text
server-management/
+-- cmd/                  (old monolith - frozen)
+-- internal/             (old monolith - frozen)
+-- microservices/        (new microservice workspace)
    +-- pkg/              (shared library code)
    +-- auth-service/
    +-- server-service/
    +-- heartbeat-gateway/
    +-- ping-worker/
    +-- monitor-service/
    +-- mail-worker/
    +-- agent/
```

## Implementation Scope Checklist

For each migration step, keep the scope limited to:

- Main source code.
- Configuration.
- API contracts.
- Message contracts.
- Dependency wiring required by the plan.

Do not include:

- Unit tests.
- Unit-test helpers.
- Opportunistic refactors outside the planned service boundary.
- Build, test, dependency, Docker, or deployment command execution unless the
  user explicitly requests it.
