# Kế hoạch di chuyển Heartbeat Gateway (`heartbeat-gateway`) làm gRPC Service

Kế hoạch này chi tiết các bước bóc tách phần tiếp nhận heartbeat thụ động từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/heartbeat-gateway`, đồng thời tận dụng các thư viện dùng chung trong `microservices/pkg` liên quan đến Kafka. 

Theo yêu cầu, service này sẽ được xây dựng làm **gRPC Service (Publisher)** thuần túy. Nó sẽ được gọi thông qua API Gateway chung thay vì trực tiếp expose RESTful HTTP handler. Điều này giúp tận dụng khả năng Native Load Balancing từ nền tảng deploy và tập trung hóa routing.

## 0. Quy tắc thực thi (Bắt buộc tuân thủ)

- **Không tự chạy các lệnh sinh code/cài đặt:** Tuyệt đối không tự chạy `go mod tidy`, không chạy `protoc` (sinh code gRPC), không build, không cài đặt dependency, không chạy Docker/deploy command. Người dùng sẽ tự thực thi các lệnh này.
- **Không tự sinh test:** Không tự sinh unit test hoặc unit-test helper code.
- **Chỉ migrate code chính:** Chỉ tập trung migrate source code chính, cấu hình (config), hợp đồng RPC (proto) và wiring cần thiết.
- **Tính chất dịch vụ:** Service này **chỉ mở cổng gRPC** để nhận request từ API Gateway nội bộ. Không kết nối Database (Postgres/Redis).
- **Theo Quyết định 6:** Implementation logic của nghiệp vụ (`Service`) phải đặt trong `internal/infra/service/`.

## 1. Tách và dùng lại Shared Packages trong `microservices/pkg`

Các package chung đã được tách ở phase trước sẽ tiếp tục được tái sử dụng:
- **`microservices/pkg/apperr`**: custom errors chung.
- **`microservices/pkg/config`**: các struct config chung.
- **`microservices/pkg/mq`**: tiện ích kết nối và publish message lên Kafka.
- **`microservices/pkg/pb/heartbeat/`**: **[NEW]** Thư mục mới để chứa `heartbeat.proto` định nghĩa hợp đồng RPC cho việc gửi heartbeat.

## 2. Xây dựng thư mục `microservices/heartbeat-gateway` (gRPC Service)

Triển khai service gRPC độc lập (ví dụ port `:50053`). Service này chịu trách nhiệm tiếp nhận heartbeat qua cổng gRPC, chuyển tiếp payload và publish vào Kafka topic `heartbeat`.

Các file/thư mục chính:
- **`cmd/main.go`**: Điểm khởi động. Wire các dependency: cấu hình, Kafka Producer, gRPC handler, và khởi chạy `grpc.NewServer()`.
- **`internal/config/config.go`**: Load cấu hình môi trường riêng cho `heartbeat-gateway` (port, kafka brokers).
- **`internal/domain/mq/publisherInterface.go`**: Interface cho việc publish message lên message queue (Kafka).
- **`internal/domain/service/gwServiceInterface.go`**: Interface cho logic nghiệp vụ tiếp nhận và publish heartbeat.
- **`internal/infra/kafka/publisher.go`**: Implement interface publisher dùng Kafka.
- **`internal/infra/service/gwService.go`**: Implement business logic của `GwServiceInterface`.
- **`internal/model/heartbeat.go`**: Struct map dữ liệu Heartbeat.
- **`internal/rpc/heartbeat_server.go`**: **[NEW]** gRPC Server handler, implement interface sinh ra từ proto, trích xuất dữ liệu từ RPC request và gọi logic từ `GwServiceInterface`.

## 3. Cấu trúc service code mục tiêu

```text
microservices/
├── pkg/
│   ├── pb/heartbeat/  # Chứa heartbeat.proto
└── heartbeat-gateway/
    ├── cmd/           # main.go (Khởi động gRPC Server port 50053)
    └── internal/
        ├── config/
        ├── domain/
        │   ├── mq/          # Chứa PublisherInterface
        │   └── service/     # Chứa GwServiceInterface
        ├── infra/
        │   ├── kafka/       # Kafka Publisher Implementation
        │   └── service/     # Implement logic nghiệp vụ GwService
        ├── model/           # Struct Heartbeat
        └── rpc/             # gRPC Handler (implement gRPC interface)
```

**Nguyên tắc Abstraction & Luồng Dependency:**
1. **Giao tiếp qua Interface:** Mọi nghiệp vụ và message queue đều phải được định nghĩa bằng Interface trong `internal/domain/`.
2. **Dependency Inversion:** Tầng logic lõi (`infra/service/gwService`) tuyệt đối không phụ thuộc vào thư viện cụ thể, mà chỉ giao tiếp qua các Interface `PublisherInterface`.
3. **gRPC Isolation:** Tầng giao tiếp (`rpc/heartbeat_server.go`) tách biệt với logic nghiệp vụ, làm nhiệm vụ trích xuất payload từ RPC request và gọi interface nghiệp vụ.
4. **Wiring Tập Trung:** Chỉ có duy nhất `cmd/main.go` là nơi inject các Implementation cụ thể vào Service.

## 4. API Contract (gRPC / Proto)

Tạo file `microservices/pkg/pb/heartbeat/heartbeat.proto`:
```protobuf
syntax = "proto3";

package heartbeat;

option go_package = "server-management/microservices/pkg/pb/heartbeat";

service HeartbeatService {
  rpc SendHeartbeat (SendHeartbeatRequest) returns (SendHeartbeatResponse);
}

message SendHeartbeatRequest {
  string server_id = 1;
  int64 timestamp  = 2;
  // Các field metadata khác từ agent nếu có
}

message SendHeartbeatResponse {
  bool success = 1;
  string message = 2;
}
```

## 5. Các bước thực thi (Execution Steps)

1. **Setup Go Module:** Đăng ký module `heartbeat-gateway` trong `go.work`.
2. **Định nghĩa Proto:** Tạo `microservices/pkg/pb/heartbeat/heartbeat.proto`.
3. **Setup Models & Domain:** Định nghĩa struct `Heartbeat` và các interface `GwServiceInterface`, `PublisherInterface`.
4. **Setup gRPC Handler:** Viết `internal/rpc/heartbeat_server.go` để xử lý gRPC request và gọi business layer.
5. **Setup Infra & Service Core:** Viết implementation gửi tin nhắn Kafka và business logic `GwService`.
6. **Bootstrap:** Tạo file `cmd/main.go` để wire tất cả dependencies và start gRPC server.
7. **Bàn giao:** Báo cáo hoàn tất cấu trúc source code cho người dùng chạy `protoc`, `go mod tidy` và test.
