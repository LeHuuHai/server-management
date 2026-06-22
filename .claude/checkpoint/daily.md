# Checkpoint Hàng Ngày (Daily Standup)

## Ngày: 22/06/2026

### 1. Trạng thái hiện tại
- Hệ thống cơ bản đã hoàn tất, **đã deploy thành công và chạy bản demo**.
- Nhận được yêu cầu nâng cấp quan trọng: Transfer toàn bộ kiến trúc sang chuẩn Microservice và tiến hành deploy quản lý bằng **Docker Swarm**.

### 2. Các công việc đang dang dở (Work In Progress - WIP)
- [ ] Phân tích các thành phần hiện tại (Master, Gateway, Worker, Writers) xem đã đủ tiêu chuẩn tách bạch Microservice chưa (về mặt state, network, config).
- [ ] Nghiên cứu và thiết kế file cấu hình Stack/Compose cho Docker Swarm.

### 3. Các công việc cần làm tiếp theo (TODO)
- [ ] Điều chỉnh source code hoặc cấu trúc thư mục (nếu cần thiết) để tuân thủ hoàn toàn mô hình Microservices độc lập.
- [ ] Viết file `docker-compose.yml` (hoặc `docker-stack.yml`) có sử dụng config `deploy` (replicas, update_config, constraints, restart_policy...) tương thích với Docker Swarm Mode.
- [ ] Đảm bảo cơ chế Service Discovery (DNS nội bộ) của Swarm hoạt động mượt mà với Kafka, Redis, ES, Postgres và các container Go.
- [ ] Test thử việc scale (scale up/down) cho Worker và Gateway qua lệnh Swarm.
- [ ] Cập nhật lại tài liệu `architecture.md` để phản ánh kiến trúc hạ tầng mới trên nền Swarm.

### 4. Vấn đề cần giải quyết / Blocker (Nếu có)
- Chạy Elasticsearch, Postgres hoặc Kafka trên Docker Swarm cần lưu ý vấn đề cấu hình mount Volumes (stateful data) vào đúng node. Cần lên phương án dùng label/constraints.
- Cơ chế quản lý cấu hình (Configs/Secrets) của Docker Swarm thay thế cho các file `.env` rời rạc hiện tại.
