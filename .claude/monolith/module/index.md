# Danh sách các Module Monolith (Legacy `cmd/`)

> **Entrypoint cho các Module của hệ thống cũ.** Thư mục này lưu trữ tài liệu phân tích sâu về vai trò, kiến trúc, và luồng dữ liệu của từng thành phần (entrypoint) trong ứng dụng Monolith.

---

## 1. Mục đích của thư mục `monolith/module/`

Tài liệu trong thư mục này giúp Agent và lập trình viên hiểu rõ chi tiết từng thành phần (process) đang chạy trong hệ thống cũ, qua đó dễ dàng định hình và chia tách code khi tiến hành refactor sang Microservices.

---

## 2. Navigation nhanh trong `monolith/module/`

| Tài liệu / Module | Vai trò chính | Thư mục source cũ |
|---|---|---|
| [`agent.md`](agent.md) | **Agent** chạy trên server mục tiêu, bắn tín hiệu heartbeat. | `cmd/agent` |
| [`gw.md`](gw.md) | **Gateway** tiếp nhận heartbeat và đẩy vào Kafka siêu tốc. | `cmd/gw` |
| [`master.md`](master.md) | **Master API:** Admin console, CRUD server, lập lịch, report. | `cmd/master` |
| [`worker.md`](worker.md) | **Worker Pool:** Ping ICMP kiểm tra uptime và gửi Email. | `cmd/worker` |
| [`pgwriter.md`](pgwriter.md) | **Batch Writer:** Ghi batch sự kiện vào PostgreSQL. | `cmd/pgwriter` |
| [`eswriter.md`](eswriter.md) | **Batch Writer:** Ghi batch sự kiện vào Elasticsearch. | `cmd/eswriter` |
