# Chi tiết các Microservices (Microservice Profiles)

Dưới đây là thông tin chi tiết về từng Microservice, bao gồm trách nhiệm, giao thức giao tiếp (APIs/RPCs), các Kafka topics liên quan, và cơ chế lưu trữ.

---

## 1. `auth-service`
*Dịch vụ quản lý tài khoản, phân quyền và cấp phát/thu hồi JWT.*

- **Port chạy:** `50051` (gRPC)
- **Cơ sở dữ liệu:**
  - **PostgreSQL:** Lưu bảng `accounts` (thông tin user, hash password).
  - **Redis:** Lưu token blocklist (các refresh token bị thu hồi khi logout hoặc token hết hạn).
- **gRPC Services:**
  - `AuthService` (đăng nhập, logout, refresh token, xác thực token).
- **Cơ chế đặc trưng:** 
  - Khi người dùng đăng xuất, Token ID được đưa vào Redis blocklist với thời gian sống (TTL) bằng thời gian hết hạn còn lại của Token để tối ưu dung lượng RAM cache.

---

## 2. `server-service`
*Dịch vụ quản lý danh mục máy chủ (Inventory Catalog). Đây là nguồn lưu trữ thông tin cấu hình tĩnh duy nhất của server.*

- **Port chạy:** `50052` (gRPC)
- **Cơ sở dữ liệu:**
  - **PostgreSQL:** Lưu bảng `servers` (id, tên, IP, version, metadata).
- **gRPC Services:**
  - `ServerService` (CRUD Server, Bulk Import XLSX, Bulk Export XLSX).
- **Kafka Topics phát hành (Publish):**
  - `server_lifecycle` (chứa các event: `ServerCreated`, `ServerUpdated`, `ServerDeleted`).
- **Cơ chế đặc trưng (Outbox & Versioning):**
  - Sử dụng **Transactional Outbox Pattern**: Việc sửa đổi database và ghi event vào bảng `outbox` được thực hiện trong cùng một Postgres transaction (qua `TxManager`). Một tiến trình background quét bảng outbox để đẩy event lên Kafka nhằm đảm bảo *At-Least-Once delivery*.
  - Mỗi server profile có trường `Version` tự tăng khi update. Event gửi đi mang theo `Version` này để consumer khử trùng và ngăn chặn việc cập nhật dữ liệu cũ đè lên dữ liệu mới.

---

## 3. `heartbeat-gateway`
*Dịch vụ nhận heartbeat với tần suất cao (high-frequency ingestion) trực tiếp từ Agent.*

- **Port chạy:** `8082` (HTTP REST)
- **Cơ sở dữ liệu:** Không có (Stateless).
- **REST API Endpoints:**
  - `POST /heartbeat` (Nhận payload: `{server_id, timestamp}`).
- **Xác thực:** Header `X-API-Key`.
- **Kafka Topics phát hành (Publish):**
  - `heartbeats` (Event: `HeartbeatReceived`).
- **Cơ chế đặc trưng:**
  - Được tối ưu hoàn toàn cho tốc độ nhận tin nhắn (< 5ms). Không kết nối DB.
  - Sử dụng Kafka Producer với cấu hình `RequiredAcks = 0` (RequireNone) để đạt thông lượng tối đa mà không bị nghẽn do chờ xác nhận từ broker.

---

## 4. `ping-worker`
*Daemon thực hiện việc ping ICMP chủ động kiểm tra trạng thái máy chủ khi heartbeat bị trễ.*

- **Port chạy:** Không mở port (Background Daemon).
- **Cơ sở dữ liệu:** Không có.
- **Kafka Topics đăng ký nhận (Consume):**
  - `ping` (Event: `PingRequested`).
- **Kafka Topics phát hành (Publish):**
  - `ping_res` (Event: `PingResult`).
- **Cơ chế đặc trưng:**
  - Cần quyền chạy hệ thống là `root` hoặc capability `CAP_NET_RAW` để mở raw-socket thực hiện ICMP ping.
  - Chạy song song nhiều instance (scale horizontally) trong cùng một Kafka consumer group để chia sẻ tải trọng ping.

---

## 5. `monitor-service`
*Trái tim của hệ thống giám sát. Quản lý trạng thái trực tuyến, nhật ký sự kiện, kiểm tra timeout và tính toán báo cáo.*

- **Port chạy:** `50054` (gRPC)
- **Cơ sở dữ liệu:**
  - **PostgreSQL:** Lưu bản sao danh sách server được giám sát (`MonitoredServer` gồm id, tên, IP, version) đồng bộ từ server-service và trạng thái trực tuyến hiện tại của server (`LiveStatus`).
  - **Elasticsearch:** Lưu nhật ký sự kiện timeseries (`StatusLog`).
  - **Redis:** Cache báo cáo ngày.
- **gRPC Services:**
  - `ReportManagementService` (RPC yêu cầu sinh báo cáo).
  - `InternalFileTransferService` (RPC stream tải file báo cáo nội bộ).
- **Kafka Topics đăng ký nhận (Consume):**
  - `server_lifecycle` (để đồng bộ danh mục server về DB Postgres local).
  - `heartbeats` (để nhận heartbeat sự kiện).
  - `ping_res` (để nhận kết quả ping chủ động).
- **Kafka Topics phát hành (Publish):**
  - `ping` (để yêu cầu ping worker kiểm tra server bị timeout).
  - `mail` (để báo cho mail-worker gửi email kèm file báo cáo).
- **Cơ chế đặc trưng:**
  - **Cơ chế quét Timeout (Active Checker):** Chạy background loop định kỳ quét trực tiếp các bảng thông tin server được giám sát (`monitoredServerRepo`) và trạng thái trực tuyến (`liveStatusRepo`) trong PostgreSQL nội bộ để xác định các server bị mất tín hiệu heartbeat nhằm phát hành yêu cầu ping.
  - **At-Least-Once Consumer:** Đọc Kafka bằng `FetchMessage` (không auto-commit). Việc commit offset chỉ thực hiện qua một closure trả về sau khi dữ liệu đã được ghi thành công xuống Postgres/Elasticsearch qua luồng Batcher.
  - **CachedAggregator:** Áp dụng MapReduce để chia nhỏ khoảng thời gian báo cáo. Các ngày đã qua được cache vĩnh viễn ở Redis, chỉ query Elasticsearch cho ngày hiện tại chưa kết thúc, giúp tăng tốc độ sinh báo cáo uptime lên 95%.

---

## 6. `mail-worker`
*Daemon gửi email báo cáo.*

- **Port chạy:** Không mở port (Background Daemon).
- **Cơ sở dữ liệu:** Không có.
- **Kafka Topics đăng ký nhận (Consume):**
  - `mail` (Event: `RequestMail`).
- **Cơ chế đặc trưng:**
  - Nhận event qua Kafka với dung lượng payload siêu nhỏ (< 1KB) chỉ chứa `filename` và thông tin người nhận để tránh nghẽn queue.
  - Khi nhận event, mail-worker tự động gọi gRPC Client kết nối đến `InternalFileTransferService` của `monitor-service` để kéo dữ liệu file nhị phân về RAM qua gRPC Stream trước khi đẩy đi bằng SMTP.

---

## 7. `agent` (Host-Level Daemon)
*Chương trình siêu nhẹ chạy trực tiếp trên các máy chủ cần giám sát.*

- **Cơ sở dữ liệu:** Không có.
- **Cơ chế hoạt động:**
  - Đọc cấu hình định danh server (`APP_SERVER_ID`) và địa chỉ heartbeat gateway.
  - Chạy vòng lặp định kỳ (cấu hình bằng mili-giây qua `APP_CYCLE_HEARTBEAT`) để gửi HTTP POST kèm API Key đến Heartbeat Gateway.
