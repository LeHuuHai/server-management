# Tên Module: ESWriter (`cmd/eswriter`)

### 1. Mục đích và Chức năng
- **Vai trò:** "Batch Writer" chuyên thu thập nhật ký (log) theo dạng chuỗi thời gian (Timeseries) và ghi vào Elasticsearch.
- **Tính năng chính:**
  - Lưu vết tất cả sự thay đổi của Heartbeat và kết quả Ping.
  - Phục vụ quá trình tính toán Report Uptime sau này.

### 2. Cấu trúc và Thành phần (Components)
- **Hạ tầng (Infra):** Kafka Consumers (x2), Elasticsearch Client (Bulk API).
- **Service Layer:** `BatchESService` thực hiện logic Micro-batching.

### 3. Luồng xử lý (Data Flow / Logic)
- Nhận event từ `heartbeat` và `ping_res` Kafka Topics.
- Map dữ liệu vào cấu trúc `model.ServerEvent` (gồm 3 field quan trọng nhất: `ServerID`, `Status`, và `Timestamp`).
- Gom đầy 1 lô (Batch size = 2000 records) trên Buffered Channel.
- Gọi hàm `WriteBatch` của Elasticsearch. Elasticsearch dùng Bulk API push mảng JSON document khổng lồ lên ES Index để Append-Only. Quá trình Indexing diễn ra cực kỳ nhanh.
