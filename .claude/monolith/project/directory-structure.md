# Cấu trúc thư mục dự án

## Root

```
server-management/
├── api/                        # OpenAPI spec + generated handlers
├── assets/                     # Ảnh diagram (Architecture.png, seq.*.png)
├── cmd/                        # Entry points — mỗi thư mục = 1 binary
├── config/                     # Per-service .env config files & structs
├── dashboard/                  # Git submodule → server-management-dashboard
├── internal/                   # Business logic (không export ra ngoài module)
├── migrations/                 # SQL migration files (golang-migrate)
├── mocks/                      # Auto-generated mocks (mockery)
├── scripts/                    # Shell scripts: init Kafka topics + ES index
├── .gitmodules
├── README.md
├── dockerfile
├── docker-compose.yaml         # Infrastructure only (PostgreSQL, Redis, Kafka, ES)
├── docker-compose.core.yaml    # Full stack
├── docker-compose.agent.yaml   # Demo agents (10 instances)
├── go.mod
├── go.sum
└── mockery.yaml
```

## `cmd/` — Service Entry Points

```
cmd/
├── agent/      # Heartbeat sender → chạy trên target server
├── eswriter/   # Batch writer → Elasticsearch
├── gw/         # Gateway → nhận heartbeat, publish Kafka
├── master/     # Main API → CRUD, auth, scheduler, cron
├── pgwriter/   # Batch writer → PostgreSQL
└── worker/     # ICMP pinger + SMTP mail sender
```

## `config/` — Per-Service Configuration

```
config/
├── agent/
│   └── .env.agent(.example)
├── common/         # Shared config types: Postgres, Redis, Kafka, ES
├── eswriter/
├── gw/
├── master/
├── pgwriter/
└── worker/
```

## `internal/` — Core Application Code

```
internal/
├── domain/                     # Contracts (interfaces) — không có implementation
│   ├── aggregator/             # ReportAggregator interface
│   ├── auth/                   # Role & Scope definitions
│   ├── cache/                  # ServerMetadataCacheInterface
│   ├── file/                   # Export & Deserialize interfaces
│   ├── mail/                   # Sender interface
│   ├── mq/                     # Publisher & Consumer interfaces
│   ├── repo/                   # Repository interfaces (server, account)
│   └── service/                # Service interfaces (server, auth, report, gw, batch, download)
│
├── handler/                    # HTTP handlers
│   ├── ServerHandler           # CRUD endpoints + import/export/report
│   ├── AuthHandler             # Login, refresh, logout
│   └── GwHandler               # Heartbeat ingestion endpoint
│
├── infra/                      # Implementations của domain interfaces
│   ├── elasticsearch/          # ES aggregator + cached aggregator + bulk writer
│   ├── file/                   # XLSX export & import (excelize)
│   ├── inmem/                  # In-memory server metadata cache (sync.Mutex map)
│   ├── jwt/                    # JWT provider (access + refresh tokens)
│   ├── kafka/                  # Kafka publisher & consumer wrappers (segmentio/kafka-go)
│   ├── mail/                   # gomail SMTP sender
│   ├── postgres/               # GORM PostgreSQL repository implementations
│   ├── redis/                  # Redis token blocklist + daily report cache
│   └── runtime/                # Per-service dependency wiring (App structs)
│
├── middleware/                 # Gin middleware
│   ├── auth/                   # JWT auth + API Key verification
│   └── logging/                # Structured request logging
│
├── model/                      # Shared data models
│   ├── Server                  # {server_id, server_name, ipv4, status, timestamps}
│   ├── Account                 # User account
│   ├── Heartbeat               # {server_id, timestamp}
│   ├── RequestPing             # {server_id, server_name, ip}
│   ├── ResponsePing            # {server_id, status, ping_at}
│   └── RequestMail             # {mail{to[], subject, attachments[]}}
│
└── service/                    # Business logic implementations
    └── batch/                  # Batch processing services (bulk write)
```

## `api/` — OpenAPI

```
api/
└── gw/         # Generated Gin handlers cho Gateway service
```

## `migrations/` — Database Schema

SQL migration files dùng `golang-migrate`. Chạy tự động khi `docker-compose.core.yaml up`.

## `scripts/`

- Init Kafka topics
- Init Elasticsearch index

## `dashboard/` (Submodule)

Frontend dashboard tại `https://github.com/LeHuuHai/server-management-dashboard`.  
Pinned commit: `164edac`.