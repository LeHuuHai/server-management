# Kiến trúc và Công nghệ (Project Architecture & Technologies)

## 1. Tổng quan thiết kế (System Architecture)
Dự án được thiết kế theo kiến trúc phân tán, hướng sự kiện (decoupled, event-driven), sử dụng Message Queue (Kafka) để tách biệt các lớp API (Web layers) với các workload xử lý nặng.

### Các Component chính (Component Roles):
- **Target Server (Agent):** Chạy `cmd/agent` để gửi heartbeats định kỳ đến Gateway.
- **Gateway (`cmd/gw`):** Stateless HTTP proxy (Gin) nhận heartbeats và publish trực tiếp vào Kafka.
- **Kafka Broker:** Đóng vai trò trung gian, chia tách lớp web và các xử lý thông qua các topics (`heartbeat`, `ping`, `ping_res`, `mail`).
- **Master API (`cmd/master`):** Console quản trị, xử lý CRUD REST APIs, xác thực/phân quyền, và chạy các cron loops (lập lịch ping, báo cáo).
- **Worker Pool (`cmd/worker`):** Các worker daemon có thể scale, thực hiện ICMP ping check chủ động và gửi email.
- **Writers (PGWriter, ESWriter):** Các service Go độc lập để nhận event từ Kafka và ghi lô (batch-write) vào PostgreSQL và Elasticsearch.
- **Databases:**
  - **PostgreSQL:** Lưu trạng thái inventory/server.
  - **Elasticsearch:** Lưu log timeseries sự kiện ping/heartbeat.
  - **Redis:** Cache metrics báo cáo tổng hợp hàng ngày & blocklist cho JWT token.

## 2. Thông tin thiết kế
- **Cấu trúc dự án (Project Structure):**
  - `/api`: OpenAPI spec và generated handlers.
  - `/cmd`: Các entrypoint của từng service (agent, eswriter, gw, master, pgwriter, worker).
  - `/config`: Cấu hình `.env` cho từng service.
  - `/internal/domain`: Chứa các interface định nghĩa core nghiệp vụ.
  - `/internal/infra`: Các triển khai hạ tầng (Postgres, Elasticsearch, Kafka, Redis, Mail...).
  - `/internal/handler`: HTTP Handlers.
  - `/internal/service`: Các Logic nghiệp vụ chính và batch service.
- **Patterns:** Clean Architecture, Message Queue / Pub-Sub (Kafka), Micro-batching.

## 3. Công nghệ sử dụng
- **Ngôn ngữ lập trình:** Go (≥ 1.21)
- **Cơ sở dữ liệu:** PostgreSQL (Relational), Elasticsearch (Timeseries/Logs), Redis (Cache)
- **Message Broker:** Kafka
- **Frameworks & Thư viện:** Gin (HTTP Routing), GORM (Postgres ORM), gomail (SMTP)
- **Công cụ triển khai:** Docker + Docker Compose (≥ 24)
