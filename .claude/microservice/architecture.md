# Kiến trúc và Cơ chế Giao tiếp (System Architecture & Communication)

Tài liệu này mô tả chi tiết thiết kế hệ thống, các luồng tương tác và cơ chế bảo mật phi tập trung (Decentralized Auth) của phiên bản Microservice.

---

## 1. Sơ đồ Kiến trúc Microservice (Architecture Diagram)

Hệ thống hoạt động theo mô hình **Hybrid Event-Driven & gRPC**:
- **Luồng dữ liệu thu thập / xử lý (Uptime Checking & Write-path):** Đi qua Kafka để đảm bảo khả năng scale lớn, chịu tải tốt, và cách ly lỗi (fault isolation).
- **Luồng điều khiển / truy vấn (Read-path & Admin):** Đi qua gRPC nội bộ để đảm bảo type-safety, tốc độ cao và đồng bộ.

```
                  ┌──────────────────────┐
                  │    Web/Mobile App    │
                  └──────────┬───────────┘
                             │ REST HTTP (JWT)
                             ▼
┌─────────────┐   REST HTTP  ┌───────────┐         gRPC (x-user-role)
│ mon-agent   │ ───────────► │  API      │ ──────────────────────────┐
│ (monitored) │   (API Key)  │  Gateway  │                           │
└─────────────┘              └─────┬─────┘                           │
                                   │                                 │
┌─────────────┐                    │ gRPC (JWT Claims)               │
│ host-agent  │ ───► POST /hb ─────┼─────────────────┐               │
│ (monitored) │ (API Key)          │                 │               │
└─────────────┘                    ▼                 ▼               ▼
                             ┌───────────┐     ┌───────────┐   ┌───────────┐
                             │   Auth    │     │ Heartbeat │   │  Server   │
                             │  Service  │     │  Gateway  │   │  Service  │
                             └─────┬─────┘     └─────┬─────┘   └─────┬─────┘
                                   │                 │               │
                                   │                 │               │ Server Lifecycle
                                   │                 │ Kafka         │ Events (Versioned)
                                   │                 ▼               ▼
┌──────────────────────────────────┼─────────────────────────────────┼─┐
│                                  │          KAFKA BUS              │ │
│  Topics:                         ▼                                 ▼ │
│  - heartbeats (Acks=0) ─────────────────────────────────────────┐    │
│  - server_lifecycle (Acks=-1) ──────────────────────────────────┼──┐ │
│  - ping (Acks=-1) ◄─────────────────────────────────────────────┼──┼─┘
│  - ping_res (Acks=-1) ───┐                                      │  │
│  - mail (Acks=-1) ◄──────┼───────────────────────┐              │  │
└──────────────────────────┼───────────────────────┼──────────────┼──┼─┘
                           │                       │              │  │
                           ▼                       │              │  │
                     ┌───────────┐                 │              │  │
                     │  Monitor  │ ◄───────────────┼──────────────┘  │
                     │  Service  │ ◄───────────────┼─────────────────┘
                     └─────┬─────┘                 │
                           │                       │
                           │ gRPC Stream           │ Kafka (mail)
                           │ (x-api-key)           │
                           ▼                       │
                     ┌───────────┐                 │
                     │   Mail    │ ◄───────────────┘
                     │  Worker   │
                     └───────────┘
```

---

## 2. Các Luồng Nghiệp Vụ Chính (Process Flows)

### Flow A: Thu thập Heartbeat thụ động (Passive Checking)
1. **Agent** gửi HTTP `POST /heartbeat` đến `heartbeat-gateway` kèm theo header `X-API-Key`.
2. **Heartbeat Gateway** xác thực API Key, đóng gói event và publish nhanh vào Kafka topic `heartbeats` (sử dụng cấu hình `RequiredAcks = 0` / fire-and-forget để tối ưu hiệu năng).
3. **Monitor Service** (chứa consumer) đọc message từ Kafka:
   - **Ghi trạng thái hiện tại:** Cập nhật `LiveStatus` vào PostgreSQL bằng cơ chế gom lô (Buffered Micro-Batch).
   - **Ghi lịch sử:** Cập nhật `StatusLog` vào Elasticsearch phục vụ báo cáo.

### Luồng phụ: Đồng bộ hóa danh mục Server (Server Synchronization)
1. Khi có bất kỳ thay đổi nào về server (thêm, sửa, xóa) tại `server-service`, một event tương ứng được gửi vào Kafka topic `server_lifecycle`.
2. **Monitor Service** đọc event này để đồng bộ và duy trì một **bản sao danh sách server** cục bộ (`MonitoredServer` table) trong PostgreSQL.
3. Bản sao cục bộ này dùng làm căn cứ danh mục máy chủ để background checker thực hiện quét và lấy thông tin địa chỉ IP.

### Flow B: Ping chủ động kiểm tra (Active Checking)
1. **Monitor Service** chạy background checker định kỳ truy vấn danh sách server được giám sát (`monitoredServerRepo`) và các trạng thái trực tuyến (`liveStatusRepo`) trực tiếp từ PostgreSQL local database. Nếu phát hiện server có thời gian nhận heartbeat gần nhất (`LastHeartbeatAt`) vượt quá timeout quy định, service sẽ gửi một event `PingRequested` vào Kafka.
2. **Ping Worker** nhận event từ topic `ping`, thực hiện ICMP ping trực tiếp tới IP của server mục tiêu.
3. Ping Worker publish kết quả `PingResult` vào Kafka topic `ping_res`.
4. **Monitor Service** đọc kết quả, cập nhật `LiveStatus` trong Postgres và ghi log Elasticsearch qua batcher.

### Flow C: Báo cáo và Gửi Mail (Report Generation & Delivery)
1. Client gọi REST API tạo báo cáo thông qua API Gateway. Gateway ủy quyền gRPC đến `ReportManagementService` trên `monitor-service`.
2. **Monitor Service** trả về HTTP `202 Accepted` ngay lập tức. Đồng thời, service truy vấn Redis cache (hoặc MapReduce ES logs) để sinh báo cáo Excel, lưu artifact cục bộ, và đẩy tin nhắn chỉ chứa `filename` vào Kafka topic `mail`.
3. **Mail Worker** nhận event từ Kafka, thực hiện gọi gRPC Client kết nối đến `InternalFileTransferService` của `monitor-service`.
4. Dữ liệu file XLSX được truyền tải an toàn bằng **gRPC Stream** qua mạng nội bộ.
5. **Mail Worker** nhận đủ stream file, tiến hành gửi email qua SMTP.

---

## 3. Kiến trúc Xác thực & Phân quyền Phi tập trung (Decentralized Security)

Để tránh nút thắt cổ chai và đảm bảo nguyên lý bảo mật Zero-Trust, hệ thống phân chia trách nhiệm xác thực như sau:

```
[REST Client] 
     │ Authorization: Bearer JWT
     ▼
┌───────────────┐
│  API Gateway  │  (Giải mã & verify JWT Token, lấy Claims)
└───────┬───────┘
        │
        │ gRPC Metadata:
        │   x-user-id: "usr_123"
        │   x-user-role: "admin"
        ▼
┌───────────────┐
│ Backend gRPC  │  (Chạy RoleCheckUnaryGRPCInterceptor so khớp role với required scope)
└───────────────┘
```

1. **Kiểm tra JWT tại API Gateway:** API Gateway là nơi duy nhất giải mã và kiểm tra chữ ký JWT từ các client bên ngoài. Sau đó, nó tự động inject các header metadata (`x-user-id`, `x-user-role`) trước khi forward request gRPC xuống các service nội bộ.
2. **Ủy quyền tại Service (Unary Interceptor):** `server-service` và `monitor-service` sử dụng `RoleCheckUnaryGRPCInterceptor` để chặn các unary RPC. Nó lấy metadata `x-user-role`, gọi hàm `Scopes()` của Role đó và so sánh xem có quyền thực thi RPC hay không. Quyền được biểu diễn bằng các struct `Scope` chặt chẽ (như `ScopeServerRead`, `ScopeServerWrite`).
3. **Mã API Key nội bộ (Streaming Interceptor):** Các luồng streaming (như `DownloadReport`) không đi qua API Gateway mà được kết nối trực tiếp bởi `mail-worker`. Luồng này được bảo vệ bằng API Key dùng chung, tự động inject ở Client qua `APIKeyBindStreamGRPCInterceptor` và được xác thực ở Server qua `APIKeyCheckStreamGRPCInterceptor` bằng metadata `x-api-key`.
4. **Agent Verification:** `heartbeat-gateway` không dùng JWT hay role, nó xác thực request gửi lên từ Agent bằng header HTTP `X-API-Key` với cấu hình cứng trong secret.
