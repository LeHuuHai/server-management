# Tên Module: Worker Pool (`cmd/worker`)

### 1. Mục đích và Chức năng
- **Vai trò:** Là các tiến trình ngầm (Daemon/Background Job worker) chạy các tác vụ tiêu tốn CPU và Network I/O nhằm "offload" (giảm tải) cho Master API. Có thể scale ngang (thêm nhiều instance Worker) một cách cực kỳ dễ dàng.
- **Tính năng chính:**
  - Thực hiện lệnh Ping ICMP đến các IP Server khi được Master phân công (Active Ping).
  - Gửi Email (SMTP) kèm file báo cáo XLSX cho quản trị viên.

### 2. Cấu trúc và Thành phần (Components)
- **Framework/Thư viện chính:** 
  - `gomail.v2`: Gửi SMTP email đính kèm file.
  - `prometheus-community/pro-bing`: Thư viện xử lý lệnh ICMP ping hệ thống.
- **Hạ tầng (Infra):** 
  - Kafka Consumer (đọc topic `ping`, `mail`).
  - Kafka Publisher (đẩy kết quả `ping_res`).
- **Service Layer:** `DownLoadService` sử dụng standard HTTP Client để gọi về Master kéo file xuống.

### 3. Luồng xử lý (Data Flow / Logic)
- **Khởi chạy (Init):** Đọc cấu hình `config/worker/.env.worker`. Khởi tạo Gomail Dialer và Kafka.
- **Tiến trình CheckServer (ICMP Pinger Worker):**
  - **Read:** Một goroutine liên tục Consume từ topic `ping`. Chuyển Kafka message thành struct `RequestPing` và đưa vào một Buffered Channel có tên là `jobs`.
  - **Process:** Khởi tạo một Goroutine Pool theo số lượng `APP_NUM_THREAD` được cấu hình. Các thread này đứng chờ tại channel `jobs`.
  - **Ping Logic:** Khi lấy được job, Worker sẽ tạo đối tượng `pinger` (Bật mode `Privileged=true` để dùng Raw Socket). Sau đó gửi duy nhất 1 gói tin ICMP với Timeout là 1s.
  - **Write:** Xác định được server `ONLINE` hay `OFFLINE`, Worker bọc kết quả vào struct `ResponsePing` và publish vào topic `ping_res` trên Kafka.
- **Tiến trình SendMail (SMTP Mailer Worker):**
  - Read topic `mail`, giải mã `RequestMail` chứa cấu trúc danh sách `Filename` đính kèm.
  - Gọi `downloadService.Download()`: Worker gọi HTTP `GET` đến endpoint của Master, truyền `APP_REPORT_KEY` để tải luồng byte của file Report XLSX xuống bộ nhớ RAM.
  - Gọi `gomailSender.Send()` để push mail qua SMTP server (Gmail/Sendgrid). Cuối cùng Commit offset Kafka.

### 4. API / Interface giao tiếp
- Consume Kafka topics: `ping`, `mail`.
- Produce Kafka topics: `ping_res`.
- Giao tiếp nội bộ (Internal HTTP): Kéo file từ Master qua `GET /report/{filename}`.
