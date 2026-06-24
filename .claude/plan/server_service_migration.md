# Kế hoạch di chuyển Server Service (`server-service`)

Kế hoạch này chi tiết các bước bóc tách phần quản lý danh mục máy chủ từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/server-service`, đồng thời tận dụng các thư viện dùng chung trong `microservices/pkg` dựa trên kiến trúc **gRPC** đã định hình ở `auth-service`.

## 0. Quy tắc thực thi (Bắt buộc tuân thủ)

- **Không tự chạy các lệnh sinh code/cài đặt:** Tuyệt đối không tự chạy `go mod tidy`, `protoc` (sinh code gRPC), `openapi-generator`, không build, không cài đặt dependency, không chạy Docker/deploy command. Người dùng sẽ tự thực thi các lệnh này.
- **Không tự sinh test:** Không tự sinh unit test hoặc unit-test helper code.
- **Chỉ migrate code chính:** Chỉ tập trung migrate source code chính, cấu hình (config), hợp đồng giao tiếp (API contract / gRPC) và wiring cần thiết.
- **Theo Quyết định 5:** Service này **chỉ mở cổng gRPC**. Tầng HTTP REST Handler sẽ được gỡ bỏ và chuyển nhiệm vụ về Gateway.
- **Theo Quyết định 6:** Implementation logic của nghiệp vụ (`Service`) phải đặt trong `internal/infra/service/`.

## 1. Tách và dùng lại Shared Packages trong `microservices/pkg`

Các package chung đã được tách ở phase trước sẽ tiếp tục được tái sử dụng:
- **`microservices/pkg/apperr`**: custom errors chung.
- **`microservices/pkg/auth`**: role/scope definitions dùng cho phân quyền. *(Lưu ý: JWT sẽ không được parse trực tiếp tại service này. Thay vào đó, API Gateway sẽ xử lý JWT và truyền `user_id`, `role` vào qua **gRPC Metadata**).*
- **`microservices/pkg/config`**: các struct config chung cho Postgres, Kafka, Redis.
- **`microservices/pkg/db`**: tiện ích mở kết nối Postgres.
- **`microservices/pkg/pb/server/`**: **[NEW]** Thư mục mới để chứa `server.proto` định nghĩa hợp đồng RPC cho quản lý máy chủ.

## 2. Xây dựng thư mục `microservices/server-service` (gRPC Server)

Triển khai service quản lý danh mục máy chủ độc lập chạy gRPC (ví dụ port `:50052`). Service này chịu trách nhiệm CRUD server, validate dữ liệu server, list/filter/sort/pagination, import/export XLSX.

Các file/thư mục chính:
- **`cmd/main.go`**: Điểm khởi động. Wire các dependency: cấu hình, Postgres, repository, Kafka Event Publisher, rpc handler, và khởi chạy `grpc.NewServer()`.
- **`internal/config/config.go`**: Load cấu hình môi trường riêng cho `server-service`.
- **`internal/domain/repo/serverRepoInterface.go`**: Interface repository cho server.
- **`internal/domain/service/serverServiceInterface.go`**: Interface cho logic nghiệp vụ lõi.
- **`internal/domain/publisher/eventPublisherInterface.go`**: **[NEW]** Interface cho việc đẩy sự kiện domain ra ngoài (tách biệt với service).
- **`internal/infra/postgres/serverRepo.go`**: Implement thao tác Postgres.
- **`internal/infra/file/*`**: Logic import/export XLSX (bóc từ tầng infra của monolith cũ).
- **`internal/infra/kafka/serverEventPublisher.go`**: **[NEW]** Implement gửi các sự kiện `ServerCreated`, `ServerUpdated`, `ServerDeleted` qua Kafka topic để `monitor-service` đồng bộ.
- **`internal/infra/runtime/rt.go`**: Khởi tạo hạ tầng (Database connection, Kafka Producer).
- **`internal/infra/service/serverService.go`**: Implement business logic của `ServerServiceInterface`.
- **`internal/model/serverProfile.go`**: Struct map dữ liệu DB.
- **`internal/rpc/server_server.go`**: gRPC Server handler, trích xuất User Context từ gRPC Metadata và gọi logic từ `ServerServiceInterface`.

## 3. Cấu trúc service code mục tiêu

```text
microservices/
├── pkg/
│   ├── pb/server/     # Chứa server.proto
└── server-service/    # Backend thuần gRPC
    ├── cmd/           # main.go (Khởi động gRPC server port 50052)
    └── internal/
        ├── config/
        ├── domain/
        │   ├── publisher/   # Chứa EventPublisherInterface
        │   ├── repo/
        │   └── service/     # Chứa ServerServiceInterface
        ├── infra/
        │   ├── file/        # Import/Export XLSX
        │   ├── kafka/       # Kafka Event Publisher
        │   ├── postgres/    # GORM Implementation
        │   ├── runtime/     # Khởi tạo DB, MQ
        │   └── service/     # Implement logic nghiệp vụ ServerService
        ├── model/           # Struct ServerProfile
        └── rpc/             # gRPC Handler (đọc gRPC Metadata)
```

**Nguyên tắc Abstraction & Luồng Dependency:**
1. **Giao tiếp qua Interface:** Mọi nghiệp vụ, data source, message queue đều phải được định nghĩa bằng Interface trong `internal/domain/`.
2. **Dependency Inversion:** Tầng logic lõi (`infra/service/serverService`) tuyệt đối không phụ thuộc vào thư viện cụ thể, mà chỉ giao tiếp qua các Interface `ServerRepoInterface`, `FileExporterInterface`, `EventPublisherInterface`.
3. **gRPC Isolation:** Tầng giao tiếp (`rpc/server_server.go`) tách biệt với logic nghiệp vụ, làm nhiệm vụ bóc tách payload, đọc gRPC Metadata, và gọi interface nghiệp vụ.
4. **Wiring Tập Trung:** Chỉ có duy nhất `cmd/main.go` là nơi inject các Implementation cụ thể (Postgres, Kafka, File) vào Service.

## 4. API Contract (gRPC) cần tách
Giao thức RPC sẽ thay thế toàn bộ REST endpoints cũ:
- `ListServers` (thay thế GET `/servers`)
- `CreateServer` (thay thế POST `/servers`)
- `UpdateServer` (thay thế PATCH `/servers/{id}`)
- `DeleteServer` (thay thế DELETE `/servers/{id}`)
- `ImportServers` (Nhận byte array trực tiếp thay thế POST `/servers/import`)
- `ExportServers` (Trả về byte array thay thế GET `/servers/export`)

*Tuyệt đối Không đưa luồng tạo báo cáo (report/download), auth, hoặc nhận heartbeat HTTP vào service này.*

## 5. Các bước thực thi (Execution Steps)

**Lưu ý Quan Trọng:** Agent sẽ KHÔNG chạy lệnh cài đặt dependency, KHÔNG chạy lệnh sinh code (`protoc`), và KHÔNG chạy `go mod tidy`. Người dùng sẽ tự thực hiện các bước này sau khi Agent bàn giao code.

1. **Định nghĩa Proto:** Tạo `microservices/pkg/pb/server/server.proto`.
2. **Setup Server-Service Models & Domain:** Copy các file model (`serverProfile.go`) và các interface (`serverServiceInterface.go`, `serverRepoInterface.go`) từ khối monolith sang cấu trúc `server-service/internal/domain/`. Tạo thêm `eventPublisherInterface.go` trong `internal/domain/publisher/`.
3. **Setup Infra & Service Core:** Copy/sửa GORM repository, file importer/exporter (xlsx), tạo implementation `serverEventPublisher.go` dùng Kafka, và chuyển `serverService.go` vào `internal/infra/service/`.
4. **Viết gRPC Handler:** Tạo `internal/rpc/server_server.go` đọc Context Metadata và gọi Business Layer.
5. **Bootstrap:** Tạo `internal/infra/runtime/rt.go` khởi tạo connections, và cập nhật `cmd/main.go` để lắng nghe TCP cho gRPC traffic.
6. **Bàn giao:** Dừng agent để người dùng tự chạy `protoc`, cài đặt module gRPC và thực thi `go mod tidy`.
