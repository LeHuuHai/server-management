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
- [x] **Transfer sang hệ thống Microservice hoàn chỉnh**
  - [x] Kế hoạch tổng thể và cấu trúc thư mục chung
  - [x] Migrate `auth-service` (REST Server)
  - [x] Migrate `server-service` (REST Server & Kafka Publisher)
  - [x] Migrate `heartbeat-gateway`
  - [x] Migrate `monitor-service` (REST Server cho API đối ngoại, gRPC cho Internal Transfer, Kafka Consumers/Publishers, Checker, Batchers, & Report generator)
  - [x] Migrate `mail-worker` và `ping-worker`
  - [x] Tích hợp HTTP Middlewares (REST) và gRPC Interceptors cho phân quyền phi tập trung (Decentralized Auth)
- [ ] **Deploy hệ thống lên hạ tầng Docker Swarm**

## 3. Các quyết định thiết kế và triển khai quan trọng (Scalability Rationales)
Dự án được xây dựng với ưu tiên cao về mặt scale (khả năng mở rộng cho hàng ngàn server).

### Quyết định 1: Tách biệt việc nhận Heartbeat và Ghi Database (Gateway & Kafka)
- **Bối cảnh:** Nếu API ghi trực tiếp heartbeat từ ngàn server vào DB ngay trong response HTTP sẽ làm chậm thread và gây I/O bottleneck.
- **Quyết định:** Dùng `heartbeat-gateway` chạy độc lập, mở cổng HTTP trực tiếp ra ngoài Internet (không đi qua API Gateway chính) để nhận HTTP request trực tiếp từ Agent và publish message vào Kafka (tốn < 5ms).
- **Lý do:** Tránh nghẽn DB và giảm tải tối đa cho API Gateway trung tâm. Việc để Agent gọi thẳng `heartbeat-gateway` giúp cô lập luồng traffic tần suất cao (high-frequency), giúp tối ưu hóa hiệu năng routing.

### Quyết định 2: Ghi dữ liệu theo lô (Buffered Micro-Batch Writing)
- **Bối cảnh:** Ghi lẻ tẻ (insert/update từng record) làm quá tải database vì overhead network và lock tranh chấp bảng.
- **Quyết định:** Tạo ra các service riêng biệt `PGWriter`, `ESWriter` đứng đọc Kafka. Chúng sẽ gom nhóm/buffer message trong RAM, khi đủ 1 batch (hoặc đủ thời gian) thì write bulk xuống DB.
- **Lý do:** Giảm tới 99% network handshakes và overhead truy vấn, giúp DB nhẹ nhàng xử lý luồng ghi khủng.

### Quyết định 3: Cache Memory & Horizontal Worker Pools cho Active Ping
- **Bối cảnh:** Quét database liên tục để tìm server bị timeout rất tốn kém (SQL read locks). Chạy hàng ngàn network ICMP Ping trực tiếp trên Master API sẽ block OS threads.
- **Quyết định:** Master API sẽ duy trì 1 `ServerInmemCache` trên RAM (Zero-Lookup Cache) để check timeout mà không cần truy vấn DB. Khi cần ping, Master ném lệnh vào Kafka. Một đàn Worker (`cmd/worker`) nhận lệnh và ping ICMP (có thể scale nhiều container worker tùy ý).
- **Lý do:** Scale ngang (Horizontal Scale) dễ dàng. Đẩy workload xử lý network I/O xuống các Worker, giữ Master API nhẹ và phản hồi nhanh.

### Quyết định 4: Cơ chế "Asynchronous 202" & "gRPC Pull" cho báo cáo
- **Bối cảnh:** Tạo báo cáo Uptime và gửi SMTP tốn nhiều thời gian. Đẩy đính kèm file XLSX to qua Kafka làm chậm broker và chiếm RAM hệ thống queue.
- **Quyết định:**
  - API Report trả về `202 Accepted` ngay lập tức để giải phóng HTTP client thread.
  - Đẩy một event `RequestMail` vào Kafka chỉ chứa `filename` (Payload cực nhỏ < 1KB).
  - Worker đọc event này, dùng gRPC Stream (thông qua `InternalFileTransferService`) để "Pull" cái file binary từ `monitor-service` rồi gửi qua gomail.
  - Sử dụng Redis để cache các kết quả `count` từ Elasticsearch (tránh query dữ liệu timeseries khổng lồ mỗi ngày).
- **Lý do:** Tối ưu kích cỡ payload Kafka, giữ tốc độ messaging cao. Giảm tải nặng cho monitor-service, truyền tải tệp tin hiệu quả và an toàn qua giao thức gRPC nội bộ thay vì HTTP công cộng.

### Quyết định 5: Sử dụng REST cho các API đối ngoại của Microservice, giữ gRPC cho giao tiếp đối nội
- **Bối cảnh:** Mặc dù gRPC mang lại hiệu năng cao và type-safety, nhưng việc chuyển đổi mọi giao tiếp client-facing sang gRPC thông qua API Gateway có thể làm phức tạp hóa quá trình tích hợp với các client bên ngoài vốn chỉ hỗ trợ REST thuần túy.
- **Quyết định:** Các dịch vụ như `auth-service`, `server-service`, và `monitor-service` sẽ trực tiếp cung cấp các REST API (sử dụng Gin và OpenAPI-codegen) cho các endpoint phục vụ client bên ngoài. API Gateway (Traefik) sẽ đơn thuần đóng vai trò proxy HTTP và xử lý JWT. Tuy nhiên, giao thức gRPC vẫn được giữ lại để xử lý **giao tiếp nội bộ** giữa các service với nhau (ví dụ: `mail-worker` stream file báo cáo từ `monitor-service`).
- **Lý do:** Tận dụng được sự tiện lợi của REST/HTTP cho các giao tiếp công cộng, đồng thời vẫn giữ được hiệu năng tuyệt đối của gRPC Protobuf cho các tác vụ truyền tải luồng dữ liệu khổng lồ trong mạng nội bộ (streaming).

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

### Quyết định 12: CachedAggregator (Decorator Pattern & MapReduce)###:
- **Bối cảnh:** Báo cáo Uptime được tạo bởi aggregator bằng cách chạy query trên ES index. Khoảng truy vấn càng dài, càng tốn tài nguyên.
- **Quyết định:** Để tối ưu truy vấn Elasticsearch khi sinh báo cáo Uptime, dịch vụ áp dụng một Decorator (`CachedAggregator`) bọc quanh base aggregator:
  1. Logic này chia nhỏ khoảng thời gian truy vấn `[from, to)` thành các phần lẻ (truy vấn trực tiếp ES) và các ngày hoàn chỉnh đã kết thúc (cache ở Redis với `TTL = 0`). Sau đó dùng cơ chế MapReduce để gộp lại theo `ServerID`, tăng tốc độ sinh báo cáo lên đáng kể và tái sử dụng bộ nhớ đệm an toàn.
  2. Report tự động hàng ngày được cache ở Redis.

### Quyết định 13: Áp dụng ISP (Interface Segregation Principle) phân tách gRPC Contract
- **Bối cảnh:** Ban đầu, gRPC contract trong monitor service gộp chung cả RPC quản lý báo cáo đối ngoại và RPC tải file đối nội. Điều này khiến các client (như `mail-worker`) phải phụ thuộc vào các RPC không liên quan.
- **Quyết định:** Chia nhỏ gRPC contract trong `monitor.proto` thành hai dịch vụ độc lập:
  1. `ReportManagementService` xử lý các yêu cầu generate report từ client/gateway bên ngoài.
  2. `InternalFileTransferService` chuyên biệt cho truyền tải file nội bộ (như RPC `DownloadReport`).
- **Lý do:** Đảm bảo nguyên lý phân tách giao diện (ISP). Mail Worker chỉ cần implement client cho dịch vụ truyền tệp (`InternalFileTransferServiceClient`), cô lập hoàn toàn khỏi logic sinh report đối ngoại, giúp hệ thống module hóa tốt hơn và bảo mật hơn.

### Quyết định 14: Kiến trúc Authentication & Authorization phi tập trung (Decentralized Auth)
- **Bối cảnh:** Việc kiểm tra xác thực (Authentication) và phân quyền (Authorization) trong hệ thống microservice dễ bị tập trung hóa quá mức tại API Gateway (khiến gateway nặng nề) hoặc bị sao chép tản mát ở các service nội bộ. Ngoài ra, cần phân biệt giữa request từ người dùng (JWT) và request nội bộ giữa các service (Service-to-Service).
- **Quyết định:** 
  1. **Authentication (Gateway):** API Gateway chịu trách nhiệm xác thực JWT token và chuyển đổi các claims thành gRPC metadata (`x-user-id`, `x-user-role`) rồi truyền tiếp xuống dưới.
  2. **Authorization (Unary RPCs):** Các service nội bộ (`server-service`, `monitor-service`) tự chịu trách nhiệm phân quyền bằng các Unary gRPC Interceptor (`AuthInterceptor` trong `pkg/auth`) dựa trên metadata role nhận được.
  3. **Service-to-Service Security (Streaming RPCs):** Các luồng gọi nội bộ không qua Gateway (ví dụ: `mail-worker` tải report từ `monitor-service`) được bảo mật bằng một mã khóa API nội bộ (`x-api-key`) xác thực qua `StreamAPIKeyInterceptor`.
  4. **Agent Ingestion Authentication (Heartbeat-Gateway):** Heartbeat gửi lên từ các remote agent được xác thực tại `heartbeat-gateway` bằng mã khóa API (`X-API-Key` HTTP Header) thay vì dùng vai trò hoặc tài khoản người dùng.
- **Lý do:** Đảm bảo tính bảo mật Zero-Trust (các service tự bảo vệ chính mình thay vì chỉ tin tưởng hoàn toàn vào Gateway), phân tách rõ ràng trách nhiệm nghiệp vụ phân quyền, và đơn giản hóa xác thực cho các dịch vụ background worker không có định danh người dùng.

### Quyết định 15: Cấu hình Triển khai Stateful Services trên Docker Swarm
- **Bối cảnh:** Khi chạy các hệ thống lưu trữ có tính trạng thái (Stateful Services) như PostgreSQL, Elasticsearch, và Kafka trên Docker Swarm, Swarm có thể tùy ý điều phối (schedule) container tới bất kỳ node nào trong cluster mỗi khi khởi động lại. Điều này làm mất kết nối tới các Local Volumes chứa dữ liệu.
- **Quyết định:** Giới hạn toàn bộ các service stateful (bao gồm cả các migration hook) chỉ được phép chạy trên một node duy nhất (Manager Node) thông qua constraint `node.role == manager` trong file `docker-stack.yml`. Sử dụng local named volumes để lưu trữ vĩnh viễn (persistent storage). 
- **Lý do:** Đây là giải pháp đơn giản và hiệu quả nhất cho kiến trúc Swarm quy mô vừa và nhỏ để tránh rủi ro phân mảnh dữ liệu (Split-brain) mà không cần cấu hình các giải pháp network storage (NFS/Ceph) phức tạp. Các stateless microservices (Worker, Gateway, v.v.) vẫn được scale ngang trên toàn bộ các Worker Nodes.

### Quyết định 16: Centralized ForwardAuth với Traefik API Gateway
- **Bối cảnh:** Mặc dù hệ thống áp dụng Decentralized Auth ở tầng gRPC nội bộ, nhưng đối với public REST API, việc để từng microservice tự parse và kiểm tra JWT HTTP Header từ Internet sẽ tạo ra lỗ hổng bảo mật và dư thừa code.
- **Quyết định:** Sử dụng middleware `ForwardAuth` của Traefik. Toàn bộ request từ Internet trước khi được routing tới các service nội bộ sẽ bị chặn lại và gửi một sub-request tới endpoint trung tâm `GET /auth/verify` thuộc `auth-service`. Nếu token hợp lệ, `auth-service` sẽ bóc tách JWT và trả về các header tùy chỉnh (`X-User-ID`, `X-User-Role`). Traefik sẽ gán các header này vào request gốc rồi chuyển tiếp (forward) xuống hạ tầng.
- **Lý do:** Tập trung hóa việc giải mã Token ở duy nhất một chốt chặn biên (API Gateway). Các microservice bên trong không cần biết JWT là gì, chỉ cần đọc `X-User-ID` từ HTTP Header tĩnh, giảm thiểu rủi ro bảo mật và triệt để tuân thủ kiến trúc API Gateway Pattern.

### Quyết định 17: Phân tách cấu hình hạ tầng và ứng dụng (Infra vs App Stack) để xử lý đồng bộ Migration
- **Bối cảnh:** Khi chạy các hệ thống Microservice trên Docker Swarm, việc ứng dụng tự động chạy schema migration lúc khởi động (trong `rt.NewApp`) có thể gây ra hiện tượng crash loop nếu các container database (Postgres, Elasticsearch) khởi động chậm và chưa sẵn sàng nhận kết nối.
- **Quyết định:** Tách cấu hình deploy thành 2 file riêng biệt: `docker-stack-infra.yml` (chứa toàn bộ Databases, Broker, Redis, API Gateway) và `docker-stack-app.yml` (chứa các microservice app). Hạ tầng (infra) sẽ được deploy trước và chờ ổn định, sau đó mới deploy tầng ứng dụng (app).
- **Lý do:** Đây là cách tiếp cận thực dụng, rõ ràng và an toàn nhất trên Docker Swarm (vốn bỏ qua cờ `depends_on`). Tránh được việc ứng dụng liên tục crash khởi động lại và hạn chế tối đa nguy cơ Data Race khi các dịch vụ cố gắng tranh chấp lock migration đồng thời.

### Quyết định 18: Loại bỏ `.env` cục bộ, sử dụng Docker Swarm Environment & Secrets
- **Bối cảnh:** Trong quá trình phát triển (Dev), thư viện `godotenv` đọc file `.env` rất tiện lợi. Nhưng khi chuyển lên Swarm, việc hardcode đường dẫn đọc file `.env` có thể gây lỗi hoặc lộ lọt thông tin cấu hình.
- **Quyết định:** Xóa hoàn toàn thư viện `godotenv` khỏi tất cả microservices. Hệ thống sẽ đọc trực tiếp từ `os.Getenv` và hàm đọc Secret từ `/run/secrets/`. Cấu hình biến môi trường sẽ được truyền động qua file `docker-stack-app.yml` và script `init-secrets.sh`.
- **Lý do:** Tuân thủ chuẩn 12-Factor App, đảm bảo môi trường Production (Swarm) luôn sạch, bảo mật qua Docker Secrets và không phụ thuộc vào các file local text không có trong image docker.
