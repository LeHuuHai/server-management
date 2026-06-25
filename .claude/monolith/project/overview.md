# Kiến trúc tổng quan — Server Management System

## 1. Mô hình kiến trúc

Hệ thống theo mô hình **distributed, event-driven** với các service độc lập giao tiếp qua Kafka message broker. Không có service nào gọi thẳng sang service khác qua HTTP (ngoại trừ Worker pull report từ Master).

```
┌─────────────┐     HTTP heartbeat      ┌─────────────┐
│  Agent      │ ──────────────────────► │  Gateway    │
│ (cmd/agent) │                         │  (cmd/gw)   │
└─────────────┘                         └──────┬──────┘
                                               │ Kafka: heartbeat
                                               ▼
┌──────────────────────────────────────────────────────────────┐
│                        KAFKA BROKER                          │
│  Topics: heartbeat │ ping │ ping_res │ mail                  │
└──────┬───────────────┬──────────────┬────────────────────────┘
       │               │              │
       ▼               ▼              ▼
┌─────────────┐  ┌──────────┐  ┌──────────────┐
│   Master    │  │  Worker  │  │  PGWriter /  │
│ (cmd/master)│  │(cmd/wkr) │  │  ESWriter    │
│             │  │          │  │              │
│ • CRUD REST │  │ • ICMP   │  │ • Batch DB   │
│ • Auth/JWT  │  │   Ping   │  │   writes     │
│ • Scheduler │  │ • SMTP   │  │              │
│ • Report    │  │   Mail   │  └──────┬───────┘
└──────┬──────┘  └──────────┘         │
       │                              ▼
       │                    ┌─────────────────────┐
       └───────────────────►│  PostgreSQL + ES     │
                            │  + Redis             │
                            └─────────────────────┘
```

## 2. Component Roles

### Target Server — Agent (`cmd/agent`)
- Chạy trên các server cần giám sát
- Gửi HTTP POST heartbeat định kỳ đến Gateway
- Payload: `{server_id, timestamp}`
- Interval cấu hình qua `APP_CYCLE_HEARTBEAT` (ms)

### Gateway (`cmd/gw`)
- Stateless HTTP proxy (Gin framework)
- Nhận heartbeat từ agents, verify API Key header (`X-API-Key`)
- Publish ngay lên Kafka topic `heartbeat` — không write DB
- Port: `APP_PORT` (default 8082)

### Kafka Broker
- Message bus trung tâm, tách biệt web layer khỏi processing layer
- Retention: 1 giờ (`KAFKA_LOG_RETENTION_MS=3600000`)
- 4 topics: `heartbeat`, `ping`, `ping_res`, `mail`

### Master API (`cmd/master`)
- Administrative console (Gin)
- CRUD REST API cho servers và accounts
- JWT authentication + Role-based authorization (Scopes)
- Cron scheduler: quét heartbeat timeout → publish ping request
- Daily report cron: midnight trigger
- Port: `APP_PORT` (default 8080)
- Swagger UI: port 8081

### Worker Pool (`cmd/worker`)
- Consume `ping` topic → thực hiện ICMP raw socket ping (timeout 1s)
- Consume `mail` topic → pull report file từ Master → gửi SMTP
- **Yêu cầu:** chạy với `root` hoặc `CAP_NET_RAW` (raw socket)
- Scale horizontally: nhiều worker instance, cùng Kafka consumer group

### PGWriter (`cmd/pgwriter`)
- Consume `heartbeat` + `ping_res` topics
- Buffer vào Go channel → bulk-write vào PostgreSQL
- Batch size: 1000 records hoặc 1-second timeout

### ESWriter (`cmd/eswriter`)
- Consume `heartbeat` + `ping_res` topics
- Buffer → bulk-write vào Elasticsearch index
- Cùng pattern với PGWriter nhưng target là ES

## 3. Ba Process Flow chính

### Flow A: Passive Uptime Checking (Heartbeat)
```
Agent → POST /heartbeat → Gateway → Kafka[heartbeat]
                                         ├→ Master (update inmem cache)
                                         ├→ PGWriter (batch → PostgreSQL)
                                         └→ ESWriter (batch → Elasticsearch)
```

### Flow B: Active Uptime Checking (ICMP Fallback)
```
Master scheduler → scan inmem cache → find timeout servers
    → publish Kafka[ping]
    → Worker → ICMP ping → publish Kafka[ping_res]
    → PGWriter + ESWriter → batch write
```

### Flow C: Report Generation & Email
```
POST /servers/report (or midnight cron)
    → Master → query Redis cache (or ES aggregation)
    → generate XLSX → save to ./tmp/
    → publish Kafka[mail] (filename only)
    → Worker → GET /report/{filename} (API Key auth) → SMTP send
    → return 202 Accepted immediately
```

## 4. Data Stores

| Store | Dùng cho |
|---|---|
| **PostgreSQL** | Inventory state (server list, accounts, last ping status) |
| **Elasticsearch** | Timeseries event logs (heartbeat + ping history) |
| **Redis** | JWT token blocklist + daily report aggregation cache |

## 5. Scalability Design Decisions

- **Gateway không write DB** → heartbeat ingestion < 5ms
- **Kafka decoupling** → zero data loss khi DB down, backpressure tự nhiên
- **In-memory cache** trên Master → scan timeout O(1) không query DB
- **Horizontal worker scaling** → thêm worker instance khi cần scale ping throughput
- **202 Accepted pattern** → report generation không block HTTP thread
- **Pull pattern cho mail** → Kafka payload nhỏ < 1KB (chỉ filename, không đính kèm file)
- **Redis aggregation cache** → tránh scan ES hàng triệu documents lặp lại