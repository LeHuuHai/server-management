# Project Monolith Structure & Services Documentation

Tài liệu này chi tiết cấu trúc thư mục của Monolith cũ để hỗ trợ lựa chọn file copy/refactor chính xác và nhanh chóng trong các bước di chuyển tiếp theo.

---

## 1. Cấu trúc thư mục Monolith (Legacy)

* **`api/`**: 
  * Chứa OpenAPI Specification (`openapi.yaml`) và code sinh ra tự động (`api.gen.go`).
  * Chỉ cần đọc để ánh xạ request/response hoặc các middleware API. **Agent không cần chạy lại lệnh code gen**.
* **`cmd/`**: 
  * Chứa các Entry Points (hàm `main`) của các ứng dụng chạy độc lập (ví dụ: `master`, `worker`, `agent`, `eswriter`, `pgwriter`, `gw`).
  * Tùy vào chức năng của microservice đích để tìm `main.go` tương ứng trong cmd cũ.
* **`config/`**: 
  * Chứa các file cấu hình và parser `.env` tương ứng với mỗi cmd (Entry Point).
* **`internal/handler/`**: 
  * Các implementation cụ thể cho router/handler phục vụ HTTP request (nhận struct API chuyển đổi và gọi qua service).
  * Chỉ cần đọc khi microservice mới có phục vụ HTTP request.
* **`internal/error/`**: 
  * Định nghĩa tập trung các custom errors dùng chung (`error.go`).
* **`internal/domain/`**: 
  * Chứa định nghĩa các interface (bản thiết kế) cho repository, service, cache, message queue.
* **`internal/infra/`**: 
  * Hiện thực chi tiết (concrete implementation) cho các interface ở domain (ví dụ: Postgres repository, Redis cache, Elasticsearch client, Kafka MQ).
* **`internal/model/`**: 
  * Định nghĩa cấu trúc dữ liệu / GORM models.
* **`internal/middleware/`**: 
  * Chứa các middleware xử lý HTTP request (xác thực JWT, phân quyền Scope, API Key cho report download, logging...).
* **`internal/service/`**: 
  * Chứa logic nghiệp vụ lõi (Business Logic) của monolith. Xem chi tiết bên dưới.

---

## 2. Chi tiết các dịch vụ trong `internal/service/`

Dưới đây là mô tả chi tiết chức năng của từng file service cũ để phục vụ tra cứu khi tạo microservices tương ứng:

| Tên File | Chức năng chi tiết trong Monolith | Microservice đích tương ứng |
| :--- | :--- | :--- |
| **`authService.go`** | Quản lý logic tài khoản, băm mật khẩu, đăng nhập, sinh token JWT, refresh token và thu hồi token (logout) qua blocklist trong Redis. | `auth-service` |
| **`serverService.go`** | Logic CRUD máy chủ, kiểm tra tính hợp lệ IP, truy vấn danh mục máy chủ với bộ lọc phân trang và sắp xếp. | `server-service` |
| **`gwService.go`** | Gateway trung gian nhận dữ liệu heartbeat thô gửi về từ Agent và thực hiện publish trực tiếp vào Kafka topic. | `heartbeat-gateway` |
| **`reportServerService.go`** | Lấy dữ liệu phân tích, gọi exporter để tạo file Excel báo cáo và bắn message qua Kafka ra lệnh gửi email báo cáo. | `monitor-service` |
| **`downloadService.go`** | Xử lý logic đọc và tải file báo cáo từ disk. | `monitor-service` |
| **`batchServerMetadataService.go`** | Service thu gom (buffer) và ghi nhận loạt (batch) thông tin heartbeat mới nhận được để cập nhật Live status vào cache. | `monitor-service` |
| **`batchPGService.go`** | Tích lũy batch trạng thái máy chủ thu thập được và cập nhật đồng loạt (Batch Update) vào Postgres DB để tối ưu hóa IO. | `monitor-service` |
| **`batchESService.go`** | Tích lũy batch các sự kiện trạng thái máy chủ thu thập được để ghi hàng loạt (Bulk Index) vào Elasticsearch nhằm lưu trữ lịch sử. | `monitor-service` |

---

## 3. Cách thức di chuyển code trong tương lai

1. **Service mới cần phục vụ HTTP (như `server-service`)**:
   * Tham khảo router cũ trong `cmd/master/main.go` và `internal/handler/serverHandler.go`.
   * Tạo route Gin trực tiếp tương tự cách làm ở `auth-service` để giữ tính độc lập.
2. **Service mới làm nhiệm vụ Worker (như `ping-worker`, `mail-worker`)**:
   * Tham khảo các cmd tương ứng trong `cmd/worker/` hoặc `cmd/pgwriter/` để copy logic khởi tạo consumer Kafka và xử lý background job.
3. **Database / Infra**:
   * Tận dụng `microservices/pkg/db` để tạo kết nối.
   * Sao chép các hàm xử lý truy vấn từ `internal/infra/postgres/` hoặc `internal/infra/redis/` tương ứng vào thư mục `internal/infra/` của service đích.
