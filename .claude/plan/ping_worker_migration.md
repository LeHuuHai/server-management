# Kế hoạch di chuyển Ping Worker (`ping-worker`)

Kế hoạch này chi tiết các bước di chuyển worker chịu trách nhiệm thực hiện ping ICMP từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/ping-worker`.

Dịch vụ này là một background worker thuần túy, không có HTTP server (trừ phi dùng cho health check/metrics sau này) và không quản lý cơ sở dữ liệu. Nó giao tiếp hoàn toàn thông qua Kafka.

## 0. Quy tắc thực thi (Bắt buộc tuân thủ)

- **Không tự chạy các lệnh sinh code/cài đặt:** Tuyệt đối không tự chạy `go mod tidy`, `protoc`, không build, không cài đặt dependency, không chạy Docker/deploy command. Người dùng sẽ tự thực thi các lệnh này.
- **Không tự sinh test:** Không tự sinh unit test hoặc unit-test helper code.
- **Chỉ migrate code chính:** Chỉ tập trung migrate source code chính, cấu hình (config), hợp đồng giao tiếp (Message contract) và wiring cần thiết.

## 1. Trách nhiệm của `ping-worker`

- Nhận lệnh Ping từ Kafka (topic `ping`).
- Thực hiện raw-socket ICMP ping (sử dụng thư viện `prometheus-community/pro-bing`).
- Đẩy kết quả trả về Kafka (topic `ping_res`).
- **Không quản lý Database**: Hoàn toàn phi trạng thái (stateless) và không lưu trữ.

## 2. Giao tiếp với các Service khác

Sử dụng cơ chế Event-Driven thông qua Kafka:
- **Tiêu thụ (Consume):** Nhận lệnh từ `monitor-service` (hoặc hệ thống khác yêu cầu ping) qua topic `ping`. Chuyển Kafka message thành struct.
- **Xuất bản (Publish):** Đẩy kết quả sau khi ping xong vào topic `ping_res` cho `monitor-service` xử lý tiếp (để tính toán trạng thái Live/StatusLog).

## 3. Cấu trúc thư mục `microservices/ping-worker`

Tuân thủ nghiêm ngặt naming conventions.

```text
microservices/
└── ping-worker/
    ├── cmd/               # main.go (Điểm khởi động, cấu hình workers pool)
    └── internal/
        ├── config/        # Load cấu hình môi trường riêng cho ping-worker
        ├── domain/
        │   ├── mq/        # Interface Consume lệnh Ping và Publish kết quả Ping
        │   └── service/   # Interface thực hiện ping (wrapper pro-bing)
        ├── infra/
        │   ├── kafka/     # Implement gửi nhận message Kafka
        │   ├── runtime/   # rt.go khởi tạo App (wiring dependencies)
        │   ├── service/   # Implement ICMP ping logic
        │   └── worker/    # Goroutine pool đọc job từ queue và gọi service
        └── model/         # Các struct RequestPing, ResponsePing
```

### Chi tiết các file chính:
- **`cmd/main.go`**: Load cấu hình, khởi tạo Kafka connections, khởi chạy Worker Pool chặn và đợi job.
- **`internal/config/config.go`**: Định nghĩa và parse các tham số cần thiết như `APP_NUM_THREAD`, cấu hình Kafka.
- **`internal/domain/mq/`**:
  - `pingConsumerInterface.go`
  - `pingResponsePublisherInterface.go`
- **`internal/domain/service/`**:
  - `pingServiceInterface.go`
- **`internal/infra/kafka/`**:
  - `pingConsumer.go`
  - `pingResponsePublisher.go`
- **`internal/infra/service/`**:
  - `pingService.go`: Thực thi logic ICMP Ping sử dụng `prometheus-community/pro-bing` (yêu cầu Privileged=true để dùng raw socket).
- **`internal/infra/worker/`**:
  - `pingWorkerPool.go`: Logic thiết lập worker pool (đọc từ channel, phân phát tới worker routines, thu kết quả).
- **`internal/model/`**:
  - `pingRequest.go` (Payload của `ping` topic)
  - `pingResponse.go` (Payload của `ping_res` topic)

## 4. Các bước thực thi (Execution Steps)

1. **Khởi tạo và Config:** Tạo thư mục `ping-worker`, thiết lập file cấu hình `internal/config/config.go`.
2. **Setup Models & Domain Interfaces:** Định nghĩa các payload model (trích xuất từ cũ), các interface cho MQ và Service.
3. **Hiện thực Infra (Kafka & Ping Service):** 
   - Viết các Publisher/Consumer tích hợp gói `microservices/pkg/mq`.
   - Viết `pingService.go` chứa logic thực thi ICMP, chuyển code cũ sang để sử dụng context và return struct gọn gàng.
4. **Xây dựng Worker Pool:** Xây dựng pool bằng Goroutine đọc Kafka message, đưa vào kênh (channel), và thực thi đồng thời. Đảm bảo offset Kafka được commit đúng cách sau khi đẩy kết quả lên topic thành công.
5. **Bootstrap (Runtime & Main):** Tạo `internal/infra/runtime/rt.go` để gom tài nguyên, và `cmd/main.go` khởi chạy ứng dụng (block the main thread).
6. **Bàn giao:** Dừng agent để người dùng tự chạy `go mod tidy` và kiểm thử.
