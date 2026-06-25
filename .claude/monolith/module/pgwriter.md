# Tên Module: PGWriter (`cmd/pgwriter`)

### 1. Mục đích và Chức năng
- **Vai trò:** "Batch Writer" - Người chuyên ghi dữ liệu theo lô cực tốc (Bulk Write) vào cơ sở dữ liệu PostgreSQL.
- **Tính năng chính:**
  - Thu thập một lượng khổng lồ các event (cả trạng thái chủ động và thụ động) từ Kafka.
  - Gom nhóm (Micro-batching) trên RAM.
  - Update toàn bộ lô xuống Postgres bằng một câu Query duy nhất, giảm cực kỳ lớn Transaction Overhead.

### 2. Cấu trúc và Thành phần (Components)
- **Hạ tầng (Infra):** Kafka Consumers (x2), PostgreSQL (`gorm`).
- **Service Layer:** Sử dụng `BatchPGService` (tự build bằng Goroutine và Channel) hỗ trợ trigger ghi theo "Kích thước lô" (Batch Size) hoặc "Thời gian trễ" (Timeout).

### 3. Luồng xử lý (Data Flow / Logic)
- **Tiến trình Consume `heartbeat`:** Lấy các message heartbeat, tạo đối tượng server với trạng thái mặc định `ONLINE`. Push vào Batch Channel.
- **Tiến trình Consume `ping_res`:** Lấy kết quả từ Worker, tạo đối tượng server với trạng thái `ONLINE`/`OFFLINE`. Push vào Batch Channel.
- **Tiến trình Gom lô (BatchService):**
  - Lấy liên tục từ Channel đưa vào một Map/Dictionary `map[string]model.Server` với key là `ServerID`. Bước này giúp ghi đè các trạng thái cũ, giữ lại trạng thái mới nhất cho mỗi ID trước khi ghi DB (Tránh việc update cùng 1 server chục lần trong 1 lô).
  - Khi lô đạt ngưỡng kích thước (`1000`) hoặc hết `1 giây`, nó gọi `serverRepo.BulkUpdateServers` của GORM để dùng `ON CONFLICT (server_id) DO UPDATE` xử lý toàn bộ.
