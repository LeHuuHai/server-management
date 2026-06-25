# Tên Module: Agent (`cmd/agent`)

### 1. Mục đích và Chức năng
- **Vai trò:** Chương trình siêu nhỏ gọn (Daemon) cài đặt tại chính các server mục tiêu (Target Servers) để đóng vai "con bệnh" cần được giám sát.
- **Tính năng chính:** Bắn tín hiệu "tôi còn sống" (Heartbeat) định kỳ về Gateway của hệ thống.

### 2. Cấu trúc và Thành phần (Components)
- **Thư viện:** Chỉ dùng Golang Standard Library `net/http` và `time`.
- **Đặc trưng:** Hoàn toàn Standalone, không lệ thuộc database, không Framework nặng. File nhị phân (binary) compile ra chiếm rất ít RAM và CPU của máy chủ mục tiêu.

### 3. Luồng xử lý (Data Flow / Logic)
- **Khởi chạy:** 
  - Nạp cấu hình từ `.env.agent` (hoặc biến môi trường), gồm: `APP_SERVER_ID` (chuỗi định danh máy chủ), URL của Gateway (`APP_HEARTBEAT_URL`), API Key và Chu kỳ bắn `APP_CYCLE_HEARTBEAT` (ví dụ 3 giây).
- **Vòng lặp sống (Live Loop):**
  - Khởi tạo một HTTP Client `http.Client` có timeout ngắt kết nối ngắn (5s).
  - Thiết lập `time.Ticker` chạy vô hạn.
  - Mỗi khi Ticker điểm nhịp, Agent gọi hàm `sendHeartbeat`:
    - Gói `ServerID` và thời gian hiện tại (`time.Now().UTC()`) thành một chuỗi JSON bytes.
    - Đính header `X-API-Key` để khai báo danh tính và POST payload lên Gateway URL.
    - Chờ Gateway phản hồi HTTP 200 OK.
