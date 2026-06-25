# Tech Stack & Dependencies

## Runtime Requirements

| Requirement | Version | Ghi chú |
|---|---|---|
| Go | ≥ 1.21 | |
| Docker + Docker Compose | ≥ 24 | Chạy infrastructure |
| Linux OS | — | ICMP ping cần `CAP_NET_RAW` |

## Infrastructure Services

| Service | Vai trò | Port (default) |
|---|---|---|
| **PostgreSQL** | Inventory state, persistent storage | 5433 |
| **Elasticsearch** | Timeseries event logs | 9200 |
| **Redis** | JWT blocklist + report aggregation cache | 6379 |
| **Kafka** | Message broker, event bus | 9092 |

## Go Libraries (key)

| Library | Dùng trong |
|---|---|
| `gin-gonic/gin` | HTTP framework (Gateway, Master) |
| `segmentio/kafka-go` | Kafka publisher & consumer |
| `gorm.io/gorm` + `gorm.io/driver/postgres` | ORM cho PostgreSQL |
| `go-redis/redis` | Redis client |
| `elastic/go-elasticsearch` | Elasticsearch client |
| `golang-jwt/jwt` | JWT access + refresh tokens |
| `xuri/excelize` | XLSX export & import |
| `gomail.v2` | SMTP email sender |
| `golang-migrate/migrate` | Database migrations |
| `vektra/mockery` | Auto-generate mocks từ interfaces |
| `go-ping/ping` | ICMP raw socket ping |

## Kafka Topics Configuration

| Topic | Retention | Producer | Consumers |
|---|---|---|---|
| `heartbeat` | 1 giờ | `gw` | `master`, `pgwriter`, `eswriter` |
| `ping` | 1 giờ | `master` | `worker` |
| `ping_res` | 1 giờ | `worker` | `pgwriter`, `eswriter` |
| `mail` | 1 giờ | `master` | `worker` |

**At-least-once delivery:** Worker chỉ commit Kafka offset cho `mail` topic sau khi SMTP send thành công.

## Authentication

- **JWT** (HMAC): Access token (TTL ~15 phút) + Refresh token (TTL ~7 ngày)
- **API Key** (`X-API-Key` header): dùng cho heartbeat endpoint (Gateway) và report download endpoint (Master)
- **Token blocklist:** Lưu trong Redis để invalidate JWT khi logout

## Authorization

Role-based access control với Scopes:

| Scope | Endpoint |
|---|---|
| `server:read` | GET /servers |
| `server:create` | POST /servers |
| `server:update` | PATCH /servers/:id |
| `server:delete` | DELETE /servers/:id |
| `server:import` | POST /servers/import |
| `server:export` | GET /servers/export |
| `server:report` | POST /servers/report |

## Docker Compose Files

| File | Mục đích |
|---|---|
| `docker-compose.core.yaml` | PostgreSQL, Redis, Kafka, ES + init migrations + Swagger |
| `docker-compose.yaml` | Full stack (core + all services) |
| `docker-compose.agent.yaml` | 10 demo agent instances |

## Code Generation

- **mockery** (`mockery.yaml`): Tự động generate mock implementations từ domain interfaces vào `mocks/`
- **OpenAPI/Swagger**: Spec tại `api/gw/openapi.yaml`, generated handlers tại `api/gw/`