# Microservices Architecture - Developer Documentation

> **Chào mừng bạn đến với tài liệu hướng dẫn phiên bản Microservice.** Đây là entrypoint chi tiết giải thích kiến trúc, cấu trúc thư mục, các service thành phần và luồng dữ liệu của phiên bản Microservice đã được di chuyển từ Monolith.

---

## 1. Bản đồ Tài liệu (Documentation Map)

Để hiểu rõ hơn về các khía cạnh của phiên bản Microservice, vui lòng xem các tài liệu sau:

| Tài liệu | Nội dung |
|---|---|
| [`architecture.md`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/.claude/microservice/architecture.md) | Kiến trúc tổng quan, mô hình giao tiếp (gRPC, Kafka, REST HTTP) và cơ chế Bảo mật/Xác thực (Decentralized Auth). |
| [`directory-structure.md`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/.claude/microservice/directory-structure.md) | Cấu trúc thư mục của workspace `microservices/` và các service thành phần. |
| [`services.md`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/.claude/microservice/services.md) | Chi tiết về 6 microservices và 1 host-level agent (API contract, Kafka topics, Database). |
| [`shared-packages.md`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/server-management/.claude/microservice/shared-packages.md) | Hướng dẫn sử dụng các thư viện dùng chung trong `microservices/pkg/` (db, mq, auth, cache, jwt, pb, v.v.). |

---

## 2. So sánh Nhanh: Monolith vs. Microservice

| Đặc tả | Phiên bản Monolith (Cũ) | Phiên bản Microservice (Mới) |
|---|---|---|
| **Cấu trúc Code** | `/cmd/` và `/internal/` ở root | `/microservices/` độc lập (Go Workspaces) |
| **Giao tiếp nội bộ** | Gọi function trực tiếp hoặc qua in-memory interfaces | gRPC (Protobuf contracts) và Kafka (Asynchronous) |
| **Bảo mật** | JWT/API Key xử lý tập trung trong Middleware của Gin | Decentralized Auth: JWT/API Key verify tại Gateway + gRPC Interceptors tại Services |
| **Database Ghi nhận** | `pgwriter`/`eswriter` chạy luồng consumer chung ở Master/Worker | `monitor-service` tích hợp các background batcher/logger đọc Kafka ghi trực tiếp |
| **Báo cáo Uptime** | Tạo file XLSX đồng bộ qua REST HTTP | Asynchronous 202: Gửi link tải qua gRPC stream + Cache Redis |
| **Triển khai** | Docker Compose thuần | Docker Swarm (Rolling updates, Secrets, Stack-level networks) |

---

## 3. Các Nguyên Tắc Vận Hành & Phát Triển (Core Guidelines)

1. **Clean Architecture tại từng Microservice:**
   Mỗi service (ví dụ `server-service`, `auth-service`) tự quản lý cấu trúc Clean Architecture của riêng nó: `domain` (contract/interface) $\rightarrow$ `infra` (implementation) $\rightarrow$ `rpc` (gRPC handlers).
2. **Không phá vỡ Interface dùng chung (Functional Options Pattern):**
   Mọi helper hoặc wrap-connection trong `pkg/` (như Kafka Writer, DB connection) phải áp dụng Functional Option Pattern để tránh breaking changes khi nâng cấp cấu hình riêng cho một service.
3. **Event-Driven & Event Ordering:**
   Các thay đổi trạng thái được truyền phát qua Kafka dưới dạng events. Sự kiện về vòng đời server được đánh version (`Version` field) để khử trùng và ngăn chặn Lost Update ở consumers.
4. **Không sử dụng `.env` ở Production:**
   Docker Swarm sẽ inject trực tiếp biến môi trường lúc deploy. Docker Secrets được sử dụng để phân phối thông tin nhạy cảm (như passwords/keys) qua đường dẫn `/run/secrets/`.
