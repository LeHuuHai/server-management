# Cấu trúc Thư mục Workspace Microservice

Thư mục `/microservices` được tổ chức theo cơ chế **Go Workspaces (`go.work`)**, cho phép quản lý nhiều microservices độc lập trong một repository duy nhất mà vẫn chia sẻ được code dùng chung.

---

## 1. Sơ đồ Tổng quan Workspace

```text
microservices/
├── go.work                       # File khai báo workspace của Go
├── go.work.sum                   # File sum chung của workspace
│
├── pkg/                          # THƯ VIỆN DÙNG CHUNG (Shared Libraries)
│   ├── apperr/                   # Định nghĩa lỗi hệ thống
│   ├── auth/                     # JWT provider & gRPC Auth/APIKey interceptors
│   ├── cache/                    # Interface memory cache
│   ├── config/                   # Config structs & log formatters
│   ├── db/                       # Khởi tạo Postgres & transaction helper (TxManager)
│   ├── es/                       # Khởi tạo Elasticsearch client
│   ├── jwt/                      # Thư viện sinh/verify JWT token
│   ├── model/                    # Domain models dùng chung (Server, Account)
│   ├── mq/                       # Wrappers cho Kafka reader & writer (kafka-go)
│   ├── rdb/                      # Khởi tạo Redis client
│   └── pb/                       # Generated Go code từ Protobuf (.proto)
│       ├── auth/                 # auth.proto & generated files
│       ├── heartbeat/            # heartbeat.proto & generated files
│       ├── server/               # server.proto & generated files
│       └── monitor/              # monitor.proto & generated files
│
├── auth-service/                 # MICROSERVICE: QUẢN LÝ TÀI KHOẢN & XÁC THỰC
│   ├── cmd/                      # Entrypoint (main.go)
│   └── internal/
│       ├── config/               # Cấu hình môi trường riêng
│       ├── domain/               # Core business interfaces (AccountRepo, JWTProvider)
│       ├── infra/                # Triển khai repository, Redis blocklist
│       │   ├── postgres/
│       │   ├── redis/
│       │   └── runtime/          # Dependency injection container (App struct)
│       ├── model/                # Entity domain model riêng
│       └── rpc/                  # Triển khai gRPC Server (AuthServiceServer)
│
├── server-service/               # MICROSERVICE: QUẢN LÝ DANH MỤC SERVER (INVENTORY)
│   ├── cmd/                      # Entrypoint (main.go)
│   └── internal/
│       ├── config/               # Cấu hình môi trường riêng
│       ├── domain/               # Core interfaces (ServerRepo, Publisher, TxManager)
│       ├── infra/                # Triển khai Postgres repo, Kafka publishers
│       │   ├── postgres/
│       │   ├── kafka/
│       │   └── runtime/          # App wiring container
│       ├── model/                # ServerProfile, OutboxEvent models
│       └── rpc/                  # Triển khai gRPC Server (ServerServiceServer)
│
├── heartbeat-gateway/            # MICROSERVICE: THU THẬP HTTP HEARTBEATS
│   ├── cmd/                      # Entrypoint (main.go)
│   ├── api/                      # REST API Contract (Generated OpenAPI code)
│   └── internal/
│       ├── config/               # Cấu hình cổng chạy và Kafka settings
│       ├── domain/               # Core interfaces (Publisher)
│       ├── handler/              # HTTP Gin handler nhận heartbeat
│       ├── infra/                # Triển khai Kafka publisher (Acks=0)
│       └── middleware/           # API Key checker middleware
│
├── monitor-service/              # MICROSERVICE: PHÂN TÍCH, GHI NHẬT KÝ & BÁO CÁO
│   ├── cmd/                      # Entrypoint (main.go)
│   └── internal/
│       ├── config/               # Cấu hình (Postgres, ES, Redis, Kafka)
│       ├── domain/               # Core interfaces (Aggregator, Checker, Cache)
│       ├── infra/                # Triển khai DB, ES query, Kafka reader/writer
│       │   ├── elasticsearch/
│       │   ├── kafka/            # Consumers & publishers
│       │   ├── postgres/         # Postgres status repos
│       │   └── redis/            # Redis cache
│       ├── model/                # LiveStatus, StatusLog, Reports models
│       └── rpc/                  # Triển khai gRPC Server (Report & FileTransfer)
│
├── ping-worker/                  # WORKER DAEMON: THỰC HIỆN ICMP PING
│   ├── cmd/                      # Entrypoint (main.go)
│   └── internal/
│       ├── config/               # Kafka consumer & producer config
│       ├── domain/               # ICMP Checker interface
│       └── infra/                # Kafka reader/writer & raw-socket pinger
│
└── mail-worker/                  # WORKER DAEMON: GỬI BÁO CÁO QUA SMTP
    ├── cmd/                      # Entrypoint (main.go)
    └── internal/
        ├── config/               # SMTP server & Kafka config
        ├── domain/               # MailSender interface
        └── infra/                # Kafka reader, SMTP driver, gRPC client pull file
```

---

## 2. Quy tắc Tổ chức Code trong từng Microservice

Mỗi microservice trong danh sách trên đều tuân theo nguyên lý thiết kế **Dependency Inversion** của Clean Architecture:

- **Tách biệt Interface và Implementation:** Interface nghiệp vụ được định nghĩa trong `internal/domain/`. Các thư viện bên ngoài hoặc driver DB (Infra layer) được định nghĩa trong `internal/infra/` và bắt buộc phải implement các interface của domain.
- **Tách biệt Config và Runtime:** File cấu hình (`internal/config/`) chỉ đọc môi trường. Việc khởi tạo các kết nối và inject dependencies được thực hiện trong `internal/infra/runtime/` (thường xuất ra một struct `App` dùng chung).
- **Lớp RPC:** Chịu trách nhiệm map dữ liệu từ protobuf struct sang domain model, thực thi service, và bắt lỗi chuyển thành gRPC codes tương ứng.
