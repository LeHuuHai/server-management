# Tên Module: Master API (`cmd/master`)

### 1. Mục đích và Chức năng
- **Vai trò:** Là module quản trị chính (Admin Console) trung tâm của toàn bộ hệ thống Server Management.
- **Tính năng chính:**
  - Cung cấp HTTP REST API qua framework Gin (CRUD server, Import/Export XLSX, Auth).
  - Xác thực (Authentication) người dùng thông qua JWT và chặn (blocklist) token qua Redis.
  - Vận hành vòng lặp (Cron loops) kiểm tra trạng thái Heartbeat của servers dựa trên In-memory Cache.
  - Lập lịch tự động tổng hợp báo cáo Uptime hàng ngày lúc nửa đêm và kích hoạt tiến trình gửi mail.

### 2. Cấu trúc và Thành phần (Components)
- **Framework/Thư viện chính:** Gin (Routing), cors, kfk (wrapper của segementio/kafka-go).
- **Domain Interfaces (`internal/domain`):** Sử dụng các interface như `ServerMetadataCacheInterface`, `mq.Publisher`, `ReportAggregator`, `ServerRepository`.
- **Hạ tầng (Infra):** Postgres (Lưu trữ metadata), Redis (Cache daily report và token blocklist), Elasticsearch (Aggregator tổng hợp buckets cho report), Kafka (như một Publisher).
- **Service Layer (`internal/service`):** Chứa các logic nghiệp vụ lõi như `ServerService`, `ReportServerService`, `AuthService`, và `BatchServerMetadataService`.

### 3. Luồng xử lý (Data Flow / Logic)
- **Khởi chạy (Init):** Đọc cấu hình từ `config/master/.env.master`. Khởi tạo Postgres, ES, Redis, Kafka Publisher và khởi chạy In-memory Server Cache bằng cách đồng bộ từ Postgres lên RAM.
- **HTTP Server (`Serve`):** Chạy API trên port được cấp, áp dụng Middleware CORS, Logger và JWT Auth.
- **Tiến trình CheckServer (Active Ping Trigger):**
  - Chạy `ticker` ngầm mỗi `APP_CYCLE_PING` ms.
  - Lấy toàn bộ danh sách server từ RAM (`ServerInmemCache`).
  - Lọc ra những server có `LastHeartbeatAt` bị quá hạn (lâu hơn `APP_HEARTBEAT_TIMEOUT`).
  - Đóng gói thông tin (IP, ServerID) thành `RequestPing` và publish vào Kafka topic `ping`.
- **Tiến trình Report:**
  - Hẹn giờ ngầm tới đúng nửa đêm (`00:00:00`).
  - Lấy kết quả Uptime aggregate từ ES/Redis, xuất báo cáo XLSX.
  - Đóng gói request gửi mail chứa "filename" vào struct `RequestMail` và publish vào Kafka topic `mail`.
- **Tiến trình ListenHeartbeat:**
  - Consume Kafka topic `heartbeat` do Gateway gửi về.
  - Đưa message vào một kênh Batch (`BatchServerMetadataService`) để gom nhóm và cập nhật lại thời gian sống (`LastHeartbeatAt`) vào RAM (In-memory Cache) mỗi giây một lần.
