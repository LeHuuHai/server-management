# Quy định và Hướng dẫn (AI Rules)

## 1. Hướng dẫn trước khi suy luận
- **BẮT BUỘC:** Đọc các file markdown trong thư mục `.claude` này trước khi suy luận, viết code hoặc thực hiện các thay đổi đối với hệ thống. Việc này giúp đảm bảo bạn có đầy đủ ngữ cảnh và tuân thủ các quy tắc của dự án.

## 2. Hướng dẫn chọn file tài liệu để đọc
Tùy thuộc vào yêu cầu của người dùng, hãy tham chiếu đến các thư mục/file sau:
- **Kiến trúc, Thiết kế và Công nghệ:** Đọc các file trong thư mục `project/`.
- **Chi tiết triển khai, Logic xử lý của từng tính năng:** Đọc các file trong thư mục `module/`.
- **Tiến độ, Công việc hàng ngày, và Các quyết định quan trọng:** Đọc các file trong thư mục `checkpoint/`.
- **Kế hoạch triển khai và Roadmap:** Đọc các file trong thư mục `plan/`.

## 3. Tổng quan dự án (Project Overview)
- **Tên dự án:** Server Management System (Hệ thống Quản lý Server)
- **Mục tiêu:** Xây dựng hệ thống phân tán, hướng sự kiện (event-driven) để theo dõi, giám sát (monitor) tình trạng sẵn sàng và các chỉ số của hệ thống server.
- **Tính năng cốt lõi (Core Features):**
  1. **Passive Health Tracking (Heartbeats):** Các server Agent tự động gửi REST HTTP heartbeats đến Gateway.
  2. **Active Health Verification (ICMP Fallback):** Ping kiểm tra ICMP chủ động do các Worker thực hiện khi heartbeats bị chậm trễ.
  3. **Timeseries Log Appending:** Lưu trữ log sự kiện vào Elasticsearch để tính toán metrics và vào PostgreSQL để lưu trạng thái.
  4. **Uptime Analytics & Reporting:** Tổng hợp báo cáo dạng file XLSX và gửi tự động hàng ngày cho Admin qua email (SMTP).
