# Phân tích Thiết kế Domain-Driven Design (DDD) cho Hệ thống Microservices

## Bước 1: Xác định Business Capabilities (Năng lực nghiệp vụ)
1. **Account & Access Management:** Đăng nhập, phân quyền, cấp JWT.
2. **Server Inventory Cataloging:** Quản lý danh mục máy chủ tĩnh.
3. **Heartbeat Ingestion:** Nhận tín hiệu nhịp tim tần suất cao.
4. **Active Uptime Probing:** Thực hiện Ping ICMP.
5. **Monitoring & Analytics:** Xử lý dữ liệu thu thập được (Cập nhật Live State, Ghi History Log, Tính toán Uptime).
6. **Notification:** Định dạng và gửi báo cáo qua Email.

## Bước 2 & 3: Các Subdomain và Bounded Contexts
Chúng ta sẽ có 6 Bounded Contexts tương ứng:

1. **Auth Context:** Domain terms (`Account`, `Token`, `Role`).
2. **Server Context:** Domain terms (`ServerProfile`, `IPAddress`).
3. **Gateway Context:** Domain terms (`RawHeartbeat`).
4. **Ping Context:** Domain terms (`PingTask`, `ICMPResult`).
5. **Monitor Context (Mới gộp):** Domain terms (`LiveStatus`, `TimeseriesEvent`, `UptimeAgg`, `ReportMetrics`).
6. **Mail Context:** Domain terms (`EmailPayload`, `Attachment`).

## Bước 4: Context Mapping (Bản đồ giao tiếp)
Sự tương tác giữa các Context được thiết kế chủ yếu theo Event-Driven (Kafka) kết hợp HTTP Pull Pattern:

- **Server Context $\rightarrow$ Monitor Context:** Bắn các sự kiện `ServerCreated`, `ServerUpdated`, `ServerDeleted` qua Kafka. Do Monitor Context lưu trữ bảng trạng thái server ở local (DB của nó), nó cần consume các event này để đồng bộ hóa danh sách, biết được server nào cần theo dõi hoặc ngừng theo dõi.
- **Gateway Context $\rightarrow$ Monitor Context:** Bắn event `HeartbeatReceived` qua Kafka.
- **Monitor Context $\rightarrow$ Ping Context:** Monitor check local cache thấy mất tín hiệu $\rightarrow$ Bắn event `PingRequested`.
- **Ping Context $\rightarrow$ Monitor Context:** Ping Worker thực hiện xong $\rightarrow$ Bắn event `PingResulted`.
- **Monitor Context $\rightarrow$ Mail Context:** Khi đến lịch, Monitor tổng hợp và sinh xong file báo cáo $\rightarrow$ Chỉ publish event `ReportAvailable` (hoặc `ReportGenerated`) để báo hiệu.
- **Mail Context $\rightarrow$ Monitor Context:** Khi Mail Context nhận được event báo hiệu, nó sẽ **chủ động gọi API (HTTP Request)** sang Monitor Context để kéo (pull) file report về và tiến hành gửi email.

---

## Bước 5: Target Architecture (6 Microservices Chốt)

Dưới đây là danh sách 6 Microservices sẽ được triển khai lên Docker Swarm:

### 1. `auth-service` (Dịch vụ Xác thực)
- **Nhiệm vụ:** API quản lý Account, verify Login, cấp JWT, chặn Token.
- **Database:** Postgres (Accounts), Redis (Blocklist).

### 2. `server-service` (Dịch vụ Danh mục Máy chủ)
- **Nhiệm vụ:** Cung cấp API Quản lý danh mục Servers (CRUD, Import/Export thông tin tĩnh).
- **Database:** Postgres (Servers).

### 3. `heartbeat-gateway` (Cổng tiếp nhận Nhịp tim)
- **Nhiệm vụ:** Proxy API chuyên nhận `POST /heartbeat` và đẩy vào Kafka siêu tốc. Không dùng DB.

### 4. `ping-worker` (Tiến trình Ping)
- **Nhiệm vụ:** Nghe Kafka nhận lệnh, thực hiện raw socket ICMP Ping, báo kết quả về Kafka. Không dùng DB.

### 5. `monitor-service` (Dịch vụ Giám sát & Phân tích)
- **Nhiệm vụ:**
  1. **Writer:** Đọc `heartbeat` & `ping_res` từ Kafka $\rightarrow$ Write Batch trạng thái hiện tại vào **Postgres** (Cũng đồng thời Consume event từ `server-service` để cập nhật bảng local).
  2. **Logger:** Đọc cùng event trên $\rightarrow$ Write Batch log lịch sử vào **Elasticsearch**.
  3. **Analyzer & Checker:** Duy trì In-memory Cache để ném lệnh Ping khi quá hạn. Đồng thời chạy Cronjob đếm log từ ES để tổng hợp Report và cung cấp API Download file cho `mail-worker`.
- **Database:** Postgres (lưu Server info và cập nhật field status), Elasticsearch, Redis (Report Cache).

### 6. `mail-worker` (Tiến trình Gửi Mail)
- **Nhiệm vụ:** Tiêu thụ lệnh gửi Mail từ Kafka, gọi API kéo file report và gửi SMTP.

---

## Bước 6: Chiến lược Tổ chức Code (Monorepo Workspace)
Toàn bộ mã nguồn của kiến trúc mới sẽ được đưa vào một thư mục root mới hoàn toàn là `microservices/` nhằm phân tách tuyệt đối với code monolith cũ.

```text
server-management/
├── cmd/                  (Monolith Cũ - ĐÓNG BĂNG)
├── internal/             (Monolith Cũ - ĐÓNG BĂNG)
└── microservices/        <=== THƯ MỤC ROOT MỚI
    ├── pkg/              (Code thư viện dùng chung)
    ├── auth-service/     
    ├── server-service/   
    ├── heartbeat-gateway/
    ├── ping-worker/
    ├── monitor-service/
    └── mail-worker/
```
