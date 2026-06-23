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
| Server Context | `ServerProfile`, `IPAddress` |
| Gateway Context | `RawHeartbeat` |
| Ping Context | `PingTask`, `ICMPResult` |
| Monitor Context | `LiveStatus`, `TimeseriesEvent`, `UptimeAgg`, `ReportMetrics` |
| Mail Context | `EmailPayload`, `Attachment` |

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
  then actively calls the `monitor-service` HTTP API to pull the report file and
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

- Expose a focused `POST /heartbeat` API.
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
  reports, cache report artifacts, and expose a download API for `mail-worker`.

Storage:

- Postgres for local server information and current status.
- Elasticsearch for historical logs and report aggregation data.
- Redis for report cache.

### 6. `mail-worker`

Responsibilities:

- Consume mail/report commands from Kafka.
- Call the `monitor-service` API to pull report files.
- Send email through SMTP.

Storage:

- No dedicated database is defined by the plan.

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
