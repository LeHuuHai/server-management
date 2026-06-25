# Kế hoạch di chuyển Mail Worker (`mail-worker`)

Kế hoạch này chi tiết các bước di chuyển worker chịu trách nhiệm gửi email kèm báo cáo (Report XLSX) từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/mail-worker`.

Dịch vụ này là một background worker thuần túy, hoạt động dựa trên thông điệp từ Kafka và gọi gRPC để lấy dữ liệu.

## 0. Quy tắc thực thi (Bắt buộc tuân thủ)

- **Không tự chạy các lệnh sinh code/cài đặt:** Tuyệt đối không tự chạy `go mod tidy`, `protoc` (sinh code gRPC), không build, không cài đặt dependency. Người dùng sẽ tự thực thi các lệnh này.
- **Không tự sinh test:** Không tự sinh unit test hoặc unit-test helper code.
- **Chỉ migrate code chính:** Tập trung migrate source code chính, cấu hình (config), kết nối (gRPC/Kafka) và logic gửi email.

## 1. Trách nhiệm của `mail-worker`

- Lắng nghe yêu cầu gửi mail từ Kafka (topic `mail`).
- Đọc thông tin các file đính kèm (`Filename`).
- Gọi sang `monitor-service` thông qua **gRPC Streaming** (`DownloadReport`) để tải nội dung file báo cáo (dạng bytes) xuống RAM.
- Gửi email thông qua giao thức SMTP (sử dụng thư viện `gomail.v2`).
- **Không quản lý Database**: Hoàn toàn phi trạng thái (stateless) và không lưu trữ file trên ổ cứng.

## 2. Giao tiếp với các Service khác

- **Với Kafka:** Consume topic `mail` để nhận `RequestMail`. Topic này thường được produce bởi `monitor-service` sau khi quá trình sinh report hoàn tất.
- **Với `monitor-service`:** Hoạt động như một gRPC Client kết nối đến `monitor-service` (ví dụ port `50054`) để stream dữ liệu file đính kèm bằng RPC `DownloadReport`.

## 3. Cấu trúc thư mục `microservices/mail-worker`

Sử dụng cấu trúc clean architecture chuẩn hóa.

```text
microservices/
└── mail-worker/
    ├── cmd/               # main.go (Điểm khởi động, kết nối gRPC, Kafka, thiết lập SMTP)
    └── internal/
        ├── config/        # Cấu hình môi trường (SMTP, Kafka, Monitor gRPC endpoint)
        ├── domain/
        │   ├── mq/        # Interface Consume lệnh Mail
        │   ├── service/   # Interface Download file (gRPC Client)
        │   └── mail/      # Interface Gửi Email (SMTP sender)
        ├── infra/
        │   ├── kafka/     # Implement nhận message từ Kafka
        │   ├── mail/      # Implement Gomail sender
        │   ├── runtime/   # rt.go khởi tạo App (wiring dependencies)
        │   ├── service/   # Implement gRPC Download Client
        │   └── worker/    # Goroutine loop đọc message, download file, gửi mail và commit
        └── model/         # Các struct RequestMail, EmailPayload, Attachment
```

### Chi tiết các thành phần chính:

- **`internal/config/config.go`**:
  - `KafkaReaderConfig` (topic: `mail`)
  - `SenderConfig` (SMTP host, port, username, password)
  - `MonitorRPC` (Địa chỉ của monitor-service gRPC server, VD: `localhost:50054`)
  - `AppConfig` (Cấu hình luồng/thời gian)

- **`internal/domain/`**:
  - `mq/mailConsumerInterface.go`: Định nghĩa hàm Read.
  - `service/downloadServiceInterface.go`: Định nghĩa hàm Download file trả về `[]byte`.
  - `mail/senderInterface.go`: Định nghĩa hàm Send email.

- **`internal/infra/service/downloadService.go`**:
  - Triển khai `DownloadServiceInterface`.
  - Khởi tạo gRPC connection (`grpc.DialContext`).
  - Gọi RPC `DownloadReport` và hứng chunk data từ stream gộp lại thành `[]byte`.

- **`internal/infra/mail/gomailSender.go`**:
  - Cấu hình Dialer với `gomail.v2`.
  - Soạn email, đính kèm bytes nội dung file (dùng `Message.Attach()`).

- **`internal/infra/worker/mailWorker.go`**:
  - Vòng lặp liên tục Consume từ Kafka.
  - Phân tách `Filename`.
  - Gọi `downloadService.Download` tải file.
  - Gọi `sender.Send` gửi mail.
  - Commit Kafka offset sau khi gửi thành công.

## 4. Các bước thực thi (Execution Steps)

1. **Khởi tạo và Config:** Tạo thư mục `mail-worker`, thiết lập file cấu hình `internal/config/config.go` trích xuất từ cấu hình cũ.
2. **Setup Models & Domain Interfaces:** Định nghĩa các payload model liên quan đến Mail (copy từ monolith sang hoặc dùng chung nếu có), tạo các giao diện MQ, Service, Mail.
3. **Hiện thực Infra (Kafka, SMTP, gRPC Client):**
   - Tạo Kafka consumer `mailConsumer.go`.
   - Tạo trình gửi mail `gomailSender.go`.
   - Tạo `downloadService.go` làm gRPC Client giao tiếp bằng proto file chung ở `pkg/pb/monitor`.
4. **Xây dựng Mail Worker:** Viết logic kết nối các luồng: lấy lệnh -> tải file -> gửi mail -> commit Kafka.
5. **Bootstrap (Runtime & Main):** Tạo `internal/infra/runtime/rt.go` để thiết lập connection pools (Kafka, gRPC, SMTP Dialer), và viết `cmd/main.go` khởi chạy ứng dụng (kèm theo Graceful Shutdown).
6. **Bàn giao:** Dừng agent để người dùng tự chạy `go mod tidy` và kiểm thử.
