# Checkpoint Hàng Ngày (Daily Standup)

## Ngày: 24/06/2026

### 1. Trạng thái hiện tại
- Hoàn thành di chuyển/migrate **`heartbeat-gateway`** và **`monitor-service`** sang kiến trúc Microservices.
- Hệ thống hỗ trợ sinh báo cáo và tải báo cáo thông qua cả cơ chế chạy định kỳ (cron-like loop) và gọi chủ động qua gRPC (`GenerateReport` và `DownloadReport`).

### 2. Các công việc đã hoàn thành trong ngày
- [x] Migrate `heartbeat-gateway` sang microservice nhận HTTP heartbeats và chuyển tiếp trực tiếp vào Kafka.
- [x] Migrate `monitor-service` hoàn chỉnh với:
  - Tách biệt Postgres lưu trữ thông tin cấu hình (`monitored_servers`) và operational status (`live_statuses`).
  - Ghi log lịch sử thay đổi trạng thái sang Elasticsearch (`StatusLog` model).
  - Tối ưu hóa ghi cơ sở dữ liệu qua micro-batching (1-second flush) cho Postgres và Elasticsearch.
  - Xây dựng active checker quét server bị quá hạn heartbeat định kỳ 5 giây.
  - Implement gRPC server hỗ trợ `DownloadReport` (stream chunks) và `GenerateReport` (gọi trực tiếp) đồng bộ với tầng nghiệp vụ sinh báo cáo (`ReportService`).
- [x] Đồng bộ hóa các trường thời gian (`MetadataUpdatedAt` đổi thành `UpdatedAt`) và loại bỏ trường thừa ở các gRPC contract và models.

### 3. Các công việc cần làm tiếp theo (TODO)
- [ ] Migrate `mail-worker` và `ping-worker`.
- [ ] Nghiên cứu và thiết kế file cấu hình Stack/Compose cho Docker Swarm.

---

## Ngày: 23/06/2026

### 1. Trạng thái hiện tại
- Hoàn tất thành công việc migrate **`server-service`** sang chuẩn Microservice.
- Đã chuẩn hóa quy tắc đặt tên (`naming_convention.md`) cho toàn bộ project.

### 2. Các công việc đã hoàn thành trong ngày
- [x] Migrate `server-service` sang gRPC server.
- [x] Cấu hình Kafka Event Publisher (phát các sự kiện `ServerCreated`, `ServerUpdated`, `ServerDeleted`).
- [x] Refactor loại bỏ In-Memory Cache khỏi `server-service` để đảm bảo Stateless Microservice.
- [x] Chuyển logic khởi tạo Kafka sang thư viện dùng chung `pkg/mq/writer.go`.
- [x] Áp dụng chuẩn **Functional Option Pattern** cho các library infrastructure (`pkg/mq`) để đảm bảo không gãy đổ hệ thống khi thêm tuỳ chọn cấu hình mới.
- [x] Triển khai **Transactional Outbox Pattern** với `TxManager` để giải quyết vấn đề Dual Write (Atomic Publishing), đảm bảo dữ liệu toàn vẹn tuyệt đối 100% giữa Postgres và Kafka.
- [x] Xử lý bài toán **Event Ordering và Idempotency (Lost Update)** theo ngữ cảnh Server:
  - Hành vi Consumer luôn là UPSERT nên hoàn toàn bỏ qua nỗi lo về Idempotency.
  - Áp dụng **Entity Versioning** cho `Server` và Event Payload để Consumer tự chặn rủi ro ghi đè ngược (Lost Update) do nhận sai thứ tự sự kiện.
- [x] Sửa lỗi Clean Architecture rò rỉ `gorm` vào tầng Service.
- [x] Viết tài liệu quy chuẩn đặt tên thư mục, file và package (`naming_convention.md`).

### 3. Các công việc cần làm tiếp theo (TODO)
- [ ] Tiến hành migrate các services tiếp theo, ưu tiên: `heartbeat-gateway` hoặc `monitor-service`.
- [ ] Nghiên cứu và thiết kế file cấu hình Stack/Compose cho Docker Swarm.
- [ ] Đảm bảo cơ chế Service Discovery (DNS nội bộ) của Swarm hoạt động mượt mà.
- [ ] Test thử việc scale cho Worker và Gateway qua lệnh Swarm.

---

## Ngày: 22/06/2026

### 1. Trạng thái hiện tại
- Hệ thống cơ bản đã hoàn tất, **đã deploy thành công và chạy bản demo**.
- Nhận được yêu cầu nâng cấp quan trọng: Transfer toàn bộ kiến trúc sang chuẩn Microservice và tiến hành deploy quản lý bằng **Docker Swarm**.

### 2. Các công việc đã hoàn thành (WIP)
- [x] Phân tích các thành phần hiện tại và định hình phương án tách Microservice.
- [x] Migrate `auth-service` sang kiến trúc gRPC.
- [x] Kế hoạch tổng thể và cấu trúc thư mục chung.

### 3. Vấn đề cần giải quyết / Blocker (Nếu có)
- Chạy Elasticsearch, Postgres hoặc Kafka trên Docker Swarm cần lưu ý vấn đề cấu hình mount Volumes (stateful data) vào đúng node. Cần lên phương án dùng label/constraints.
- Cơ chế quản lý cấu hình (Configs/Secrets) của Docker Swarm thay thế cho các file `.env` rời rạc hiện tại.
