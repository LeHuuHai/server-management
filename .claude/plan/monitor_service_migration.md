# Kế hoạch di chuyển Monitor Service (`monitor-service`)

Kế hoạch này chi tiết các bước di chuyển các logic liên quan đến giám sát trạng thái hệ thống, phân tích dữ liệu, ghi log lịch sử và tạo báo cáo từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/monitor-service`.

Dịch vụ này sử dụng cơ chế Event-Driven thông qua Kafka làm luồng giao tiếp chính, đồng thời mở cổng gRPC để hỗ trợ file-transfer (tải báo cáo) thông qua cơ chế streaming.

## 0. Quy tắc thực thi (Bắt buộc tuân thủ)

- **Không tự chạy các lệnh sinh code/cài đặt:** Tuyệt đối không tự chạy `go mod tidy`, `protoc` (sinh code gRPC), không build, không cài đặt dependency, không chạy Docker/deploy command. Người dùng sẽ tự thực thi các lệnh này.
- **Không tự sinh test:** Không tự sinh unit test hoặc unit-test helper code.
- **Chỉ migrate code chính:** Chỉ tập trung migrate source code chính, cấu hình (config), hợp đồng giao tiếp (API contract / gRPC) và wiring cần thiết.
- **Theo Quyết định 5:** Các internal microservices giao tiếp trực tiếp qua gRPC. API tải báo cáo sẽ được triển khai dạng **gRPC Stream** để truyền file an toàn và tiết kiệm bộ nhớ.
- **Theo Quyết định 6:** Implementation logic của nghiệp vụ (`Service`) phải đặt trong `internal/infra/service/`.

## 1. Tách và dùng lại Shared Packages trong `microservices/pkg`

Các package chung đã được tách ở các phase trước sẽ được tái sử dụng:
- **`microservices/pkg/apperr`**: custom errors chung.
- **`microservices/pkg/config`**: các struct config chung cho Postgres, Kafka, Elasticsearch, Redis.
- **`microservices/pkg/db`**: tiện ích mở kết nối Postgres.
- **`microservices/pkg/mq`**: tiện ích kết nối và publish/consume message lên Kafka.
- **`microservices/pkg/pb/monitor/`**: **[NEW]** Thư mục mới để chứa `monitor.proto` định nghĩa hợp đồng RPC hỗ trợ streaming tải file báo cáo.

## 2. Xây dựng thư mục `microservices/monitor-service`

Triển khai dịch vụ `monitor-service` chạy song song cả gRPC Server (ví dụ port `:50054`) và các luồng background worker đọc Kafka.

Các file/thư mục chính:
- **`cmd/main.go`**: Điểm khởi động. Khởi tạo cấu hình, kết nối Database (Postgres, Elasticsearch, Redis), khởi chạy gRPC Server, và chạy các Kafka Consumer Workers dưới nền.
- **`internal/config/config.go`**: Load cấu hình môi trường riêng cho `monitor-service` (PG, ES, Redis, Kafka Brokers/Topics).
- **`internal/domain/repo/`**: Interface repository cho dữ liệu metadata máy chủ (`monitoredServerRepoInterface.go`), trạng thái operational (`liveStatusRepoInterface.go`), và logs.
- **`internal/domain/mq/`**: Các Interface phục vụ gửi và nhận message từ Kafka.
- **`internal/domain/service/`**: Các Interface cho logic nghiệp vụ:
  - `monitorServiceInterface.go`: Cập nhật trạng thái, xử lý heartbeat/ping.
  - `reportServiceInterface.go`: Tạo báo cáo, thống kê uptime.
- **`internal/infra/postgres/`**: Gorm repositories để cập nhật thông tin servers (`monitoredServerRepo.go`) và live status của servers (`liveStatusRepo.go`).
- **`internal/infra/elasticsearch/`**: Ghi log lịch sử trạng thái dạng Bulk.
- **`internal/infra/redis/`**: Cache các báo cáo hoặc tracking keys.
- **`internal/infra/kafka/`**: Implement gửi nhận message (Ping request, Mail request, etc.).
- **`internal/infra/service/`**: Implement logic nghiệp vụ:
  - `batchPGService.go`: Gom heartbeat/ping result và batch update Postgres (chu kỳ flush 1s).
  - `batchESService.go`: Gom logs và bulk ghi xuống Elasticsearch (chu kỳ flush 1s).
  - `monitorService.go`: Xử lý logic cập nhật trạng thái chung.
  - `reportService.go`: Tổng hợp dữ liệu từ ES, ghi file Excel báo cáo.
- **`internal/infra/worker/`**: Các background thread:
  - `lifecycleConsumer.go`: Consume server events từ `server-service` để cập nhật bảng server nội bộ.
  - `heartbeatConsumer.go`: Consume heartbeat từ `heartbeat-gateway`.
  - `pingResultConsumer.go`: Consume kết quả ping từ `ping-worker`.
  - `activeChecker.go`: Chạy chu kỳ 5s quét kiểm tra các server quá hạn heartbeat để gửi yêu cầu Ping qua Kafka.
- **`internal/rpc/monitor_server.go`**: Implement gRPC server, cung cấp stream tải file báo cáo Excel.

## 3. Cấu trúc service code mục tiêu

```text
microservices/
├── pkg/
│   ├── pb/monitor/        # Chứa monitor.proto
└── monitor-service/
    ├── cmd/               # main.go (gRPC server port 50054 & workers)
    └── internal/
        ├── config/
        ├── domain/
        │   ├── aggregator/  # Chứa ReportAggregator
        │   ├── cache/       # Chứa DailyReportCacheInterface
        │   ├── mq/
        │   ├── repo/
        │   └── service/
        ├── infra/
        │   ├── elasticsearch/ # ES Client / Writer
        │   ├── kafka/         # Kafka Publisher/Consumer helpers
        │   ├── postgres/      # GORM repositories
        │   ├── redis/         # Redis Client
        │   ├── runtime/       # rt.go khởi tạo mọi DB/Clients/Connections
        │   ├── service/       # Implement logic (Gom Batch)
        │   └── worker/        # Các background thread (Checker 5s, Consumers)
        ├── model/             # Structs (LiveStatus, MonitoredServer, StatusLog)
        └── rpc/               # gRPC Stream download handler
```

## 4. API Contract (gRPC / Proto)

Tạo file `microservices/pkg/pb/monitor/monitor.proto`:
```protobuf
syntax = "proto3";

package monitor;

option go_package = "github.com/LeHuuHai/server-management/microservices/pkg/pb/monitor;monitorpb";

service MonitorService {
  rpc DownloadReport (DownloadReportRequest) returns (stream DownloadReportResponse);
}

message DownloadReportRequest {
  string filename = 1;
}

message DownloadReportResponse {
  bytes chunk_data = 1;
}
```

## 5. Các bước thực thi (Execution Steps)

1. **Định nghĩa Proto:** Tạo `microservices/pkg/pb/monitor/monitor.proto`.
2. **Setup Models & Domain:** Di chuyển/định nghĩa các model và interface cốt lõi cho Repository, MQ, và Services.
3. **Setup Infra & Service Core:**
   - Viết các dịch vụ gom batch (BatchPG, BatchES) kế thừa từ monolith với cấu hình chu kỳ flush 1s.
   - Viết logic kết nối Elasticsearch, Redis, Postgres.
4. **Setup Workers:**
   - Cài đặt `activeChecker.go` chạy chu kỳ 5s so sánh thời gian heartbeat.
   - Cài đặt các Consumer để hứng server events, heartbeat và ping results.
5. **Setup gRPC Handler:** Viết `internal/rpc/monitor_server.go` hỗ trợ stream dữ liệu chunk-by-chunk.
6. **Bootstrap:** Tạo `internal/infra/runtime/rt.go` để gom toàn bộ kết nối tài nguyên, và `cmd/main.go` khởi chạy.
7. **Bàn giao:** Dừng agent để người dùng tự sinh code gRPC, chạy `go mod tidy` và kiểm thử.

## 6. Các quyết định kiến trúc cốt lõi (Cập nhật sau triển khai)
1. **At-Least-Once Delivery (Closure-based Commits)**:
   Các Kafka consumer (Heartbeat, Lifecycle, PingResponse) được thiết kế ở chế độ Stateless. Chúng sử dụng `FetchMessage` (thay vì auto-commit) và trả về một closure `commitFunc func(context.Context) error`. Worker background chỉ gọi hàm commit này sau khi logic nghiệp vụ đã được thực thi thành công, qua đó bảo toàn Metadata của Kafka (Topic, Partition, Offset) mà không bị rò rỉ lên tầng Domain.
2. **CachedAggregator (Decorator Pattern & MapReduce)**:
   Để tối ưu truy vấn Elasticsearch khi sinh báo cáo Uptime, dịch vụ áp dụng một Decorator (`CachedAggregator`) bọc quanh base aggregator. Logic này chia nhỏ khoảng thời gian truy vấn `[from, to)` thành các phần lẻ (truy vấn trực tiếp ES) và các ngày hoàn chỉnh đã kết thúc (cache ở Redis với `TTL = 0`). Sau đó dùng cơ chế MapReduce để gộp lại theo `ServerID`, tăng tốc độ sinh báo cáo lên đáng kể và tái sử dụng bộ nhớ đệm an toàn.
