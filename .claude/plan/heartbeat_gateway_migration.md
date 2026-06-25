# Kế hoạch di chuyển Heartbeat Gateway (`heartbeat-gateway`) làm HTTP Service

Kế hoạch này chi tiết các bước bóc tách phần tiếp nhận heartbeat thụ động từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/heartbeat-gateway`, đồng thời tận dụng các thư viện dùng chung trong `microservices/pkg` liên quan đến Kafka. 

Theo yêu cầu mới, service này sẽ được xây dựng làm **HTTP REST Service** tiếp nhận trực tiếp traffic gửi heartbeat từ các remote Host Agent. Điều này giúp tối ưu hóa hiệu năng, giảm tải cho API Gateway trung tâm, và xử lý tập trung bằng mã khóa API (`X-API-Key`).

## 0. Quy tắc thực thi (Bắt buộc tuân thủ)

- **Không tự chạy các lệnh sinh code/cài đặt:** Tuyệt đối không tự chạy `go mod tidy`, không chạy `protoc` (nếu không cần thiết), không build, không cài đặt dependency, không chạy Docker/deploy command. Người dùng sẽ tự thực thi các lệnh này.
- **Không tự sinh test:** Không tự sinh unit test hoặc unit-test helper code.
- **Chỉ migrate code chính:** Chỉ tập trung migrate source code chính, cấu hình (config), HTTP handler, và wiring cần thiết.
- **Tính chất dịch vụ:** Service này **chỉ mở cổng HTTP** để nhận request trực tiếp từ Agent. Không kết nối Database (Postgres/Redis).
- **Theo Quyết định 6:** Implementation logic của nghiệp vụ (`Service`) phải đặt trong `internal/infra/service/`.

## 1. Tách và dùng lại Shared Packages trong `microservices/pkg`

Các package chung đã được tách ở phase trước sẽ tiếp tục được tái sử dụng:
- **`microservices/pkg/apperr`**: custom errors chung.
- **`microservices/pkg/config`**: các struct config chung.
- **`microservices/pkg/mq`**: tiện ích kết nối và publish message lên Kafka.

## 2. Xây dựng thư mục `microservices/heartbeat-gateway` (HTTP Service)

Triển khai service HTTP độc lập (ví dụ port `:8082`). Service này chịu trách nhiệm tiếp nhận heartbeat qua cổng HTTP POST `/heartbeat`, xác thực qua API Key, chuyển tiếp payload và publish vào Kafka topic `heartbeat`.

Các file/thư mục chính:
- **`cmd/main.go`**: Điểm khởi động. Wire các dependency: cấu hình, Kafka Producer, HTTP router, và khởi chạy HTTP Server.
- **`internal/config/config.go`**: Load cấu hình môi trường riêng cho `heartbeat-gateway` (port, kafka brokers, API Key).
- **`internal/domain/mq/publisherInterface.go`**: Interface cho việc publish message lên message queue (Kafka).
- **`internal/domain/service/gwServiceInterface.go`**: Interface cho logic nghiệp vụ tiếp nhận và publish heartbeat.
- **`internal/infra/kafka/publisher.go`**: Implement interface publisher dùng Kafka.
- **`internal/infra/service/gwService.go`**: Implement business logic của `GwServiceInterface`.
- **`internal/model/heartbeat.go`**: Struct map dữ liệu Heartbeat.
- **`internal/middleware/apiKeyMiddleware.go`**: Middleware xác thực API Key của Agent qua header `X-API-Key`.
- **`internal/handler/httpHandler.go`**: HTTP Server handler, tiếp nhận request JSON và gọi logic từ `GwServiceInterface`.

## 3. Cấu trúc service code mục tiêu

```text
microservices/
└── heartbeat-gateway/
    ├── cmd/           # main.go (Khởi động HTTP Server port 8082)
    └── internal/
        ├── config/
        ├── domain/
        │   ├── mq/          # Chứa PublisherInterface
        │   └── service/     # Chứa GwServiceInterface
        ├── handler/         # HTTP Handler nhận JSON request
        ├── infra/
        │   ├── kafka/       # Kafka Publisher Implementation
        │   └── service/     # Implement logic nghiệp vụ GwService
        ├── middleware/      # Middleware xác thực API Key
        └── model/           # Struct Heartbeat
```

**Nguyên tắc Abstraction & Luồng Dependency:**
1. **Giao tiếp qua Interface:** Mọi nghiệp vụ và message queue đều phải được định nghĩa bằng Interface trong `internal/domain/`.
2. **Dependency Inversion:** Tầng logic lõi (`infra/service/gwService`) tuyệt đối không phụ thuộc vào thư viện cụ thể, mà chỉ giao tiếp qua các Interface `PublisherInterface`.
3. **HTTP Isolation:** Tầng giao tiếp (`handler/httpHandler.go`) tách biệt với logic nghiệp vụ, làm nhiệm vụ trích xuất payload từ HTTP request và gọi interface nghiệp vụ.
4. **Wiring Tập Trung:** Chỉ có duy nhất `cmd/main.go` là nơi inject các Implementation cụ thể vào Service.

## 4. API Contract (HTTP)

Endpoint: `POST /heartbeat`

Headers:
- `Content-Type: application/json`
- `X-API-Key: <valid_api_key>`

Request Body:
```json
{
  "server_id": "string"
}
```

Response:
- Success: `202 Accepted`
- Unauthorized: `401 Unauthorized` (nếu thiếu hoặc sai API Key)
- Bad Request: `400 Bad Request` (nếu thiếu hoặc sai cấu trúc payload)

## 5. Các bước thực thi (Execution Steps)

1. **Setup Config:** Cập nhật `internal/config/config.go` để load thêm cấu hình API Key từ environment.
2. **Setup Middleware & Handler:** Viết `apiKeyMiddleware.go` và `httpHandler.go` cho máy chủ HTTP.
3. **Dọn dẹp code gRPC:** Xóa thư mục/file `internal/rpc/` và gỡ bỏ dependencies liên quan đến protobuf/gRPC server trong `cmd/main.go`.
4. **Bootstrap HTTP Server:** Viết lại `cmd/main.go` để khởi chạy Gin/HTTP server thay vì gRPC server.
5. **Bàn giao:** Báo cáo hoàn tất cấu trúc source code cho người dùng chạy `go mod tidy` và test.
