# Tên Module: Gateway (`cmd/gw`)

### 1. Mục đích và Chức năng
- **Vai trò:** Là proxy tiếp nhận (Ingress) thụ động chuyên đảm nhận việc xử lý lưu lượng Heartbeat khổng lồ từ các Agent.
- **Tính năng chính:**
  - Nhận HTTP POST Request từ hàng ngàn Agent.
  - Xác thực siêu tốc bằng `X-API-Key` Middleware.
  - Đẩy event vào Kafka và trả về HTTP response ngay lập tức (Thời gian phản hồi mục tiêu < 5ms).

### 2. Cấu trúc và Thành phần (Components)
- **Framework/Thư viện chính:** Gin (chạy mode nhẹ nhất), Kafka Publisher.
- **Hạ tầng (Infra):** Chỉ kết nối tới một hạ tầng duy nhất là Kafka Broker. Không kết nối DB (Postgres/Redis) để tránh block pool kết nối.
- **Service Layer:** `GwService` là một lớp rất mỏng, chỉ nhận dữ liệu và gọi Kafka Publisher.

### 3. Luồng xử lý (Data Flow / Logic)
- **Khởi chạy (Init):** Đọc cấu hình `config/gw/.env.gw`. Cấu hình duy nhất 1 connection pool tới Kafka.
- **HTTP Server (`Serve`):** Khởi động Gin server kèm `APIKeyMiddleware`.
- **Tiếp nhận Request:**
  - Khi `/heartbeat` được gọi, Request body JSON sẽ được giải mã ra struct `Heartbeat`.
  - Service gọi `Publisher.Publish()` để đẩy message vào Kafka topic `heartbeat`.
  - Không có xử lý nghiệp vụ hay I/O nào khác diễn ra. Server trả về HTTP 200 ngay lập tức để Agent ngắt connection, giải phóng port cho HTTP Thread.

### 4. API / Interface giao tiếp
- REST API: Mở endpoint public `POST /heartbeat` cho Agent gọi.
- Producer: Kafka topic `heartbeat`.
