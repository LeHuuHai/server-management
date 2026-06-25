# Tiến độ dự án và Các quyết định quan trọng

## 1. Các phần đã hoàn thành
- Setup dự án và chia tách kiến trúc các services (Gateway, Master, Worker, Agent, Writers).
- Triển khai thành công luồng Active & Passive Checking (ICMP Ping & HTTP Heartbeat).
- Lưu trữ decoupling thông qua Kafka, ghi log ra Elasticsearch và Postgres.
- Cài đặt hệ thống Cron xuất báo cáo XLSX và gửi qua Email tự động.
- **[Cập nhật mới] Dự án đã được deploy và chạy demo thành công.**

## 2. Tiến độ tổng thể
- [x] Kiến trúc Event-Driven với Kafka
- [x] Master CRUD API cho Servers (JWT & Roles)
- [x] Bulk Import/Export Excel
- [x] Worker ICMP Pinger & Mail Dispatch
- [x] Deploy và chạy Demo hệ thống hiện tại
- [ ] **Transfer sang hệ thống Microservice hoàn chỉnh**
  - [x] Kế hoạch tổng thể và cấu trúc thư mục chung
  - [x] Migrate `auth-service` (gRPC Server)
  - [x] Migrate `server-service` (gRPC Server & Kafka Publisher)
  - [x] Migrate `heartbeat-gateway`
  - [x] Migrate `monitor-service` (gRPC Server with DownloadReport/GenerateReport, Kafka Consumers/Publishers, Checker, Batchers, & Report generator)
  - [ ] Migrate `mail-worker` và `ping-worker`
- [ ] **Deploy hệ thống lên hạ tầng Docker Swarm**

## 3. Các quyết định thiết kế và triển khai quan trọng (Scalability Rationales)
Dự án được xây dựng với ưu tiên cao về mặt scale (khả năng mở rộng cho hàng ngàn server).

### Quyết định 1: Tách biệt việc nhận Heartbeat và Ghi Database (Gateway & Kafka)
- **Bối cảnh:** Nếu API ghi trực tiếp heartbeat từ ngàn server vào DB ngay trong response HTTP sẽ làm chậm thread và gây I/O bottleneck.
- **Quyết định:** Dùng Gateway chỉ nhận HTTP request và publish message vào Kafka (tốn < 5ms).
- **Lý do:** Tránh nghẽn DB. Kafka đóng vai trò "giảm xóc" (Backpressure). Dù Postgres hay Elasticsearch chậm lại, Gateway cũng không sập và dữ liệu không bị mất.

### Quyết định 2: Ghi dữ liệu theo lô (Buffered Micro-Batch Writing)
- **Bối cảnh:** Ghi lẻ tẻ (insert/update từng record) làm quá tải database vì overhead network và lock tranh chấp bảng.
- **Quyết định:** Tạo ra các service riêng biệt `PGWriter`, `ESWriter` đứng đọc Kafka. Chúng sẽ gom nhóm/buffer message trong RAM, khi đủ 1 batch (hoặc đủ thời gian) thì write bulk xuống DB.
- **Lý do:** Giảm tới 99% network handshakes và overhead truy vấn, giúp DB nhẹ nhàng xử lý luồng ghi khủng.

### Quyết định 3: Cache Memory & Horizontal Worker Pools cho Active Ping
- **Bối cảnh:** Quét database liên tục để tìm server bị timeout rất tốn kém (SQL read locks). Chạy hàng ngàn network ICMP Ping trực tiếp trên Master API sẽ block OS threads.
- **Quyết định:** Master API sẽ duy trì 1 `ServerInmemCache` trên RAM (Zero-Lookup Cache) để check timeout mà không cần truy vấn DB. Khi cần ping, Master ném lệnh vào Kafka. Một đàn Worker (`cmd/worker`) nhận lệnh và ping ICMP (có thể scale nhiều container worker tùy ý).
- **Lý do:** Scale ngang (Horizontal Scale) dễ dàng. Đẩy workload xử lý network I/O xuống các Worker, giữ Master API nhẹ và phản hồi nhanh.

### Quyết định 4: Cơ chế "Asynchronous 202" & "HTTP Pull" cho báo cáo
- **Bối cảnh:** Tạo báo cáo Uptime và gửi SMTP tốn nhiều thời gian. Đẩy đính kèm file XLSX to qua Kafka làm chậm broker và chiếm RAM hệ thống queue.
- **Quyết định:**
  - API Report trả về `202 Accepted` ngay lập tức để giải phóng HTTP client thread.
  - Đẩy một event `RequestMail` vào Kafka chỉ chứa `filename` (Payload cực nhỏ < 1KB).
  - Worker đọc event này, dùng HTTP để "Pull" cái file binary từ Master API rồi gửi qua gomail.
  - Sử dụng Redis để cache các kết quả `count` từ Elasticsearch (tránh query dữ liệu timeseries khổng lồ mỗi ngày).
- **Lý do:** Tối ưu kích cỡ payload Kafka, giữ tốc độ messaging cao. Giảm tải nặng cho Master API.

### Quyết định 5: Chuyển đổi REST sang gRPC cho các Microservices nội bộ
- **Bối cảnh:** Khi tách hệ thống thành các Microservices độc lập (ví dụ: `auth-service`), việc giao tiếp qua HTTP/REST sẽ cồng kềnh, thiếu type-safety và chậm hơn.
- **Quyết định:** Các Backend Microservices chỉ mở cổng gRPC. API Gateway sẽ là component duy nhất nhận HTTP/REST từ Client bên ngoài và proxy thành request gRPC vào hệ thống nội bộ.
- **Lý do:** Tối ưu hóa hiệu năng mạng, sử dụng Protobuf làm strict contract, tách biệt tầng routing HTTP ra khỏi logic service.

### Quyết định 6: Cấu trúc thư mục Dependency Inversion (Service nằm trong Infra)
- **Bối cảnh:** Cần duy trì tính decoupling tuyệt đối giữa logic nghiệp vụ và cách nó được khởi tạo.
- **Quyết định:** Chuyển phần implementation của `service` (ví dụ `AuthService`) vào thư mục `internal/infra/service/`. Interface vẫn ở `internal/domain/`. Bootstrap layer ở `cmd/main.go` sẽ phụ trách inject implementation.
- **Lý do:** Đảm bảo kiến trúc Clean Architecture hoàn toàn tuân thủ Dependency Inversion Principle, tách bạch "việc cấu hình/khởi tạo" ra khỏi "logic lõi".

### Quyết định 7: Sử dụng Functional Option Pattern cho Shared Infrastructure
- **Bối cảnh:** Khi xây dựng các thư viện dùng chung cho nhiều microservices (`pkg/mq`, `pkg/db`), mỗi service có thể cần những cấu hình khác nhau (ví dụ: sync vs async, batch size, timeout). Nếu thêm tham số trực tiếp vào hàm khởi tạo sẽ làm phá vỡ interface của các service cũ.
- **Quyết định:** Áp dụng **Functional Options Pattern** (ví dụ: `opts ...WriterOption`) cho các constructor dùng chung trong thư mục `microservices/pkg/`.
- **Lý do:** Đảm bảo tính tương thích ngược (backward compatibility) 100%. Các service có thể dễ dàng override cấu hình mặc định (như `WithAsync(true)`) mà không ảnh hưởng tới code của các service khác. Giữ cho việc khởi tạo luôn sạch sẽ và có thể mở rộng vô hạn.

### Quyết định 8: Quản lý Transaction qua Context (TxManager Pattern)
- **Bối cảnh:** Khi áp dụng Transactional Outbox Pattern, thao tác ghi `Server` và ghi `Outbox Event` phải nằm chung trong một Database Transaction. Tuy nhiên, nếu truyền trực tiếp `*gorm.DB` vào các interface ở tầng `Domain/Service` thì sẽ phá vỡ quy tắc Clean Architecture (tầng Domain không được biết về công nghệ Infra).
- **Quyết định:** Xây dựng `TxManagerInterface` với hàm `WithTx(ctx context.Context, fn func(txCtx context.Context) error)`. Object Transaction được inject ngầm (ẩn) vào bên trong `context.Context` tại tầng Infra. Các Repository (`ServerRepo`, `OutboxRepo`) tự động kiểm tra `context` để lấy transaction ra sử dụng chung.
- **Lý do:** Tầng `Service` (Business logic) hoàn toàn "mù" về Gorm hay Postgres, đảm bảo Decoupling tuyệt đối 100%. Code ở Service cực kỳ gọn gàng nhưng vẫn đạt được Data Consistency tối đa.

### Quyết định 9: Đảm bảo Event Ordering và Ngăn chặn Lost Update
- **Bối cảnh:** Khi gửi event qua Kafka, Consumer có thể nhận trùng event (duplicate) hoặc nhận sai thứ tự (Out-of-order).
- **Quyết định:** 
  1. **Idempotency:** Trong ngữ cảnh quản lý Server, hành vi của Consumer (`monitor-service`) luôn là cập nhật trạng thái (UPSERT). Do đó, chúng ta **không cần quan tâm đến vấn đề Idempotency**. Việc nhận trùng event nhiều lần không gây ra side-effect.
  2. **Ordering & Lost Update:** Để tránh việc Consumer lấy dữ liệu cũ (đến trễ) ghi đè lên dữ liệu mới (Lost Update), ta áp dụng **Entity Versioning**. Thêm trường `Version` vào model `Server`, mỗi lần cập nhật sẽ tăng `Version + 1` và đính kèm vào Event.
- **Lý do:** Cách tiếp cận cực kỳ thực dụng. Mọi Consumer chỉ cần thực hiện câu truy vấn `UPSERT ... WHERE version < incoming_version` để tự động chặn các event tới sai thứ tự.

### Quyết định 10: Điều chỉnh Acknowledgment (RequiredAcks) theo tính chất dữ liệu
- **Bối cảnh:** Các microservice khác nhau giao tiếp qua Kafka có các yêu cầu khác nhau về độ tin cậy của dữ liệu (Consistency vs Availability/Performance).
- **Quyết định:**
  1. **Đối với `server-service` (Dữ liệu cấu hình hệ thống):** Cấu hình `RequiredAcks = -1` (`RequireAll`). Writer sẽ đợi phản hồi ghi nhận từ toàn bộ replica (In-Sync Replicas). Tránh mất mát dữ liệu vòng đời Server, kết hợp với Entity Versioning ở Consumer để khử trùng (Idempotency).
  2. **Đối với `heartbeat-gateway` (Dữ liệu metric thời gian thực):** Cấu hình `RequiredAcks = 0` (`RequireNone`). Writer gửi tin nhắn theo dạng fire-and-forget mà không đợi phản hồi từ broker. Dữ liệu heartbeat có tính chất tần suất cao và tạm thời (ephemeral), mất mát một vài tin nhắn không ảnh hưởng tới tính đúng đắn chung của hệ thống.
- **Lý do:** Tối ưu hóa hiệu năng và thông lượng (throughput) cho dịch vụ thu thập heartbeat (`heartbeat-gateway`) đạt độ trễ cực thấp (< 1ms), trong khi bảo toàn tính nhất quán mạnh mẽ cho cấu hình server (`server-service`).

### Quyết định 11: Đảm bảo At-Least-Once Delivery và Stateless Consumer bằng Closure
- **Bối cảnh:** Mặc định `kafka-go` sử dụng `ReadMessage` sẽ tự động commit offset ngay sau khi đọc. Nếu worker crash trước khi xử lý xong (lưu DB), message sẽ bị mất (At-Most-Once). Mặt khác, nếu thêm hàm `Commit()` vào interface của Domain, Consumer ở tầng Infra sẽ phải lưu trữ `kafka.Message` dưới dạng biến trạng thái (Stateful) - gây rủi ro Data Race khi chạy đa luồng, hoặc vi phạm Clean Architecture do rò rỉ struct của Kafka lên Domain.
- **Quyết định:** Sử dụng `FetchMessage` (không tự động commit) và thay đổi hàm `Read()` để trả về một closure `commitFunc func(context.Context) error`. Worker sẽ chủ động gọi `commitFunc` sau khi nghiệp vụ thành công.
- **Lý do:** Đảm bảo **At-Least-Once Delivery** (không bao giờ mất message khi crash). Closure giúp bảo toàn toàn bộ Metadata của Kafka (Topic, Partition, Offset) mà không làm rò rỉ chúng lên tầng Domain. Quan trọng nhất, cách tiếp cận này giữ cho Consumer struct hoàn toàn **Stateless**, an toàn tuyệt đối khi scale ngang các worker goroutine.
