# Thư viện Dùng chung (Shared Packages - `/pkg`)

Thư mục `microservices/pkg/` chứa các thư viện dùng chung cho toàn bộ các microservices. Việc tập trung hóa các thư viện này giúp duy trì sự nhất quán về mặt công nghệ, giảm thiểu trùng lặp code và chuẩn hóa các giao thức kết nối.

---

## 1. Danh sách các Shared Packages

| Thư mục | Chức năng | Các thành phần chính |
|---|---|---|
| [`apperr`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/apperr) | Định nghĩa lỗi nghiệp vụ | Danh sách mã lỗi chuẩn hóa, chuyển đổi từ Domain Error sang gRPC status codes. |
| [`auth`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/auth) | Xác thực & Phân quyền gRPC | `RoleCheckUnaryGRPCInterceptor` (Unary auth server), `APIKeyCheckStreamGRPCInterceptor` (Stream auth server), `APIKeyBindStreamGRPCInterceptor` (Stream auth client). Định nghĩa `Role` & `Scope`. |
| [`cache`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/cache) | Kết nối Redis (cơ bản) | Hàm khởi tạo Redis client không bọc lỗi chuẩn hóa. |
| [`config`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/config) | Định nghĩa Struct Cấu hình | Các struct config cho DB, Redis, Kafka, ES. Hỗ trợ log formatters qua `slog.LogValue`. |
| [`db`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/db) | Database Driver & Transactions | Khởi tạo Postgres (GORM), định nghĩa `TxManager` hỗ trợ transactional outbox pattern. |
| [`es`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/es) | Elasticsearch Connection | Wrapper khởi tạo ES client và cấu hình bulk operations. |
| [`jwt`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/jwt) | Cấp phát & Verify JWT | Sinh Access Token, Refresh Token và phân tích claims. |
| [`model`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/model) | Shared Domain Entities | Model `Server` (versioned) và `Account` dùng chung giữa các DB schemas. |
| [`mq`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/mq) | Message Queue wrappers | Kafka Reader & Writer wrappers sử dụng `segmentio/kafka-go` hỗ trợ manual offset commits. |
| [`pb`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/pb) | Generated Protobuf Code | Các struct và service clients biên dịch từ gRPC `.proto` contracts. |
| [`rdb`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/microservices/pkg/rdb) | Kết nối Redis (chuẩn hóa lỗi) | Khởi tạo Redis client và map lỗi kết nối thông qua `apperr`. |

---

## 2. Các Pattern Thiết kế Quan trọng áp dụng tại `/pkg`

### Pattern 1: Functional Options Pattern cho Initialization
Để đảm bảo code khởi tạo dùng chung linh hoạt và tương thích ngược (không làm crash các service khác khi một service cần thêm tuỳ chỉnh cấu hình), tất cả các hàm constructor trong `pkg/` đều sử dụng Functional Options.

**Ví dụ Kafka Writer Initialization:**
```go
// Khởi tạo mặc định
writer := mq.NewKafkaWriter(cfg)

// Khởi tạo tùy chỉnh riêng cho heartbeat-gateway (Acks=0 cho tốc độ cao)
writer := mq.NewKafkaWriter(cfg, mq.WithRequiredAcks(0), mq.WithAsync(true))
```

### Pattern 2: Quản lý Transaction qua Context (`TxManager`)
Để hỗ trợ Transactional Outbox Pattern ở `server-service` mà không làm rò rỉ thư viện hạ tầng (GORM/PostgreSQL) lên tầng nghiệp vụ (Domain Layer), hệ thống sử dụng một interface `TxManager` truyền qua `context.Context`.

- Tầng **Service** chỉ cần gọi `TxManager.WithTx(ctx, func(txCtx) error { ... })` để bắt đầu transaction.
- Tầng **Repository** tự động trích xuất transaction ngầm từ `txCtx` để thực thi câu lệnh SQL. Nếu có lỗi, transaction tự rollback; ngược lại tự commit.
- Tầng **Domain** hoàn toàn không chứa bất cứ import nào liên quan đến SQL hay GORM.

### Pattern 3: Stateless Consumer với Manual Offset Commit
Mặc định thư viện kafka-go sẽ commit offset ngay sau khi đọc (At-Most-Once). Để chuyển thành **At-Least-Once Delivery** (không mất message nếu service bị crash giữa chừng), wrapper `KafkaReader` được thiết kế để trả về message kèm theo một closure commit:

```go
msg, commitFunc, err := reader.FetchMessage(ctx)
if err != nil {
    return err
}

// Thực thi logic ghi DB/ES thành công...
err = processMessage(msg)
if err == nil {
    commitFunc(ctx) // Chỉ commit khi xử lý nghiệp vụ thành công
}
```
Nhờ đóng gói metadata (topic, partition, offset) trong closure `commitFunc`, struct của Consumer hoàn toàn không lưu trạng thái (Stateless), đảm bảo an toàn tuyệt đối trước vấn đề race condition khi xử lý đa luồng (multi-goroutines).
