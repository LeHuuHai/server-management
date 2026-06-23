# Microservice Migration Plan

> **Entrypoint cho quá trình di chuyển (migration) sang Microservice.** Thư mục này chứa các tài liệu quy hoạch, hướng dẫn và kế hoạch chi tiết để chuyển đổi hệ thống từ Monolith sang Microservices.

---

## 1. Mục đích của thư mục `plan/`

Thư mục này đóng vai trò là "bản thiết kế" cho việc tách các chức năng từ ứng dụng Monolith cũ thành các Microservices độc lập. Mọi tác vụ bóc tách, refactor hay tạo service mới đều phải tuân thủ các quy định và kế hoạch được định nghĩa tại đây.

---

## 2. Navigation nhanh trong `plan/`

Dưới đây là danh sách các tài liệu liên quan và thứ tự ưu tiên khi đọc:

| Tài liệu | Nội dung / Vai trò |
|---|---|
| [`microservice_migration_guide.md`](microservice_migration_guide.md) | **Đọc đầu tiên.** Quy tắc chung, kiến trúc tổng thể của Microservice, phân chia Bounded Contexts và danh sách 6 services mục tiêu. |
| [`monolith_structure.md`](monolith_structure.md) | Phân tích cấu trúc thư mục của dự án Monolith cũ. Cần đọc để biết tìm code cũ ở đâu khi tiến hành di chuyển. |
| [`auth_service_migration.md`](auth_service_migration.md) | Kế hoạch di chuyển chi tiết cho **Auth Service** (xử lý đăng nhập, JWT, quản lý tài khoản). |
| [`server_service_migration.md`](server_service_migration.md) | Kế hoạch di chuyển chi tiết cho **Server Service** (quản lý danh mục máy chủ, CRUD, import/export). |

---

## 3. Quy tắc chung khi làm việc với `plan/`

1. **Tuân thủ Migration Rules:** Luôn đọc và tuân thủ các "Migration Execution Rules" trong file `microservice_migration_guide.md`. Không tự ý chạy các lệnh thay đổi environment (như `go mod tidy`, sinh code, build, v.v.) trừ khi được yêu cầu.
2. **Migration theo Phase:** Các file `*_migration.md` mô tả kế hoạch cho từng service cụ thể. Hãy tham chiếu vào tài liệu của service tương ứng khi thực hiện công việc.
3. **Giữ nguyên codebase cũ:** Code của monolith cũ (`cmd/`, `internal/`) được xem là "frozen". Các thay đổi sẽ tập trung vào thư mục `microservices/` mới.
