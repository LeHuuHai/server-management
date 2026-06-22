# Server Management System — Agent Entrypoint

> **Đọc file này đầu tiên.** Đây là entrypoint định hướng ngữ cảnh, rule và đường dẫn tài liệu liên quan cho mọi agent làm việc trên dự án này.

---

## 1. Tổng quan dự án

**Server Management System** là một hệ thống giám sát hạ tầng phân tán, hướng sự kiện (event-driven), theo dõi tính khả dụng và metrics của các server fleet.

- **Ngôn ngữ chính:** Go (≥ 1.21) — 98.7% codebase
- **Repo:** https://github.com/LeHuuHai/server-management
- **Branch chính:** `master`
- **Dashboard (submodule):** https://github.com/LeHuuHai/server-management-dashboard

---

## 2. Tính năng cốt lõi

| Tính năng | Mô tả |
|---|---|
| Passive Health Tracking | Host server đẩy heartbeat HTTP đến Gateway |
| Active Health Verification | ICMP ping fallback khi heartbeat bị trễ |
| Timeseries Logging | Ghi sự kiện vào Elasticsearch + PostgreSQL |
| Uptime Reporting | Tạo file XLSX và gửi email hàng ngày qua SMTP |

---

## 3. Rule cho Agent

1. **Luôn đọc file liên quan** trước khi thay đổi một module — xem mục 5 bên dưới để biết đường dẫn.
2. **Tuân thủ clean architecture**: domain → infra → handler, không import ngược.
3. **Interface trước implementation**: mọi logic nghiệp vụ phải có interface trong `internal/domain/`.
4. **Kafka là bus chính** giữa các service — không gọi thẳng DB từ Gateway hay Agent.
5. **In-memory cache** (`ServerInmemCache`) là nguồn truth cho trạng thái real-time, DB là persistent state.

---

## 4. Cấu trúc thư mục `.claude/`

```
.claude/
├── index.md                  ← FILE NÀY (đọc đầu tiên)
├── architecture/
│   ├── overview.md           ← Kiến trúc tổng quan, component diagram
│   ├── directory-structure.md← Cấu trúc thư mục chi tiết
│   └── tech-stack.md         ← Công nghệ và dependencies
├── modules/
│   ├── agent.md              ← Module: cmd/agent
│   ├── gateway.md            ← Module: cmd/gw
│   ├── master.md             ← Module: cmd/master
│   ├── worker.md             ← Module: cmd/worker
│   ├── pgwriter.md           ← Module: cmd/pgwriter
│   ├── eswriter.md           ← Module: cmd/eswriter
│   ├── domain.md             ← internal/domain (interfaces)
│   └── infra.md              ← internal/infra (implementations)
├── plan/                     ← Kế hoạch migrate sang microservice 
└── checkpoint/               ← Checkpoint tiến độ migrate 
```

---

## 5. Navigation nhanh

| Cần tìm hiểu về... | Đọc file |
|---|---|
| Kiến trúc tổng quan & flow | `architecture/overview.md` |
| Cấu trúc thư mục dự án | `architecture/directory-structure.md` |
| Công nghệ, thư viện, dependencies | `architecture/tech-stack.md` |
| Agent (heartbeat sender) | `modules/agent.md` |
| Gateway (HTTP ingress) | `modules/gateway.md` |
| Master API (CRUD + scheduler) | `modules/master.md` |
| Worker (ICMP pinger + mailer) | `modules/worker.md` |
| PGWriter / ESWriter | `modules/pgwriter.md`, `modules/eswriter.md` |
| Domain interfaces (contracts) | `modules/domain.md` |
| Infrastructure implementations | `modules/infra.md` |
| Kế hoạch microservice migration | `plan/`  |
| Tiến độ migration hiện tại | `checkpoint/`  |

---

## 6. Kafka Topics tóm tắt

| Topic | Producer | Consumers |
|---|---|---|
| `heartbeat` | `gw` | `master`, `pgwriter`, `eswriter` |
| `ping` | `master` | `worker` |
| `ping_res` | `worker` | `pgwriter`, `eswriter` |
| `mail` | `master` | `worker` |

---

## 7. API Endpoints nhanh

Base URL (Master): `http://localhost:8080`  
Auth: `Authorization: Bearer <access_token>` (JWT)

| Method | Path | Mô tả |
|---|---|---|
| POST | `/auth/login` | Đăng nhập, lấy JWT |
| GET | `/servers` | Danh sách servers |
| POST | `/servers` | Tạo server |
| PATCH | `/servers/:id` | Cập nhật server |
| DELETE | `/servers/:id` | Xóa server |
| POST | `/servers/import` | Import XLSX |
| GET | `/servers/export` | Export XLSX |
| POST | `/servers/report` | Tạo report (async 202) |
| GET | `/report/:filename` | Download report (API Key) |

Swagger UI: `http://localhost:8081`

---

## 8. Ghi chú migrate (đọc thêm plan/ và checkpoint/)

Dự án đang trong quá trình **migrate sang microservice architecture**. Kế hoạch và tiến độ được lưu tại:
- `plan/` — Kế hoạch tổng thể, phase breakdown, dependency map
- `checkpoint/` — Trạng thái hiện tại, những gì đã xong, những gì đang làm
