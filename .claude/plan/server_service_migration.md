# Kế hoạch di chuyển Server Service (`server-service`)

Kế hoạch này chi tiết các bước bóc tách phần quản lý danh mục máy chủ từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/server-service`, đồng thời tận dụng các thư viện dùng chung trong `microservices/pkg` dựa trên cấu trúc đã định hình ở `auth-service`.

## 0. Quy tắc thực thi

- Không chạy bất kỳ project command nào trong quá trình migration, trừ khi người dùng yêu cầu rõ ràng.
- Không chạy `go mod tidy`, không gen API code, không chạy test, không build, không install/update dependency, không chạy Docker/deploy command.
- Không tự sinh unit test hoặc unit-test helper code.
- Chỉ migrate source code chính, config, API contract và wiring cần thiết.
- Nếu có command nên chạy để kiểm tra, chỉ ghi chú lại như manual follow-up cho người dùng.

## 1. Tách và dùng lại Shared Packages trong `microservices/pkg`

Các package chung đã được tách cho `auth-service` cần được dùng lại ở `server-service` để tránh duplicate logic:

- **`microservices/pkg/apperr`**: custom errors chung. Nguồn gốc từ `internal/error/error.go`.
- **`microservices/pkg/auth`**: role/scope definitions dùng cho authorization.
- **`microservices/pkg/jwt`**: verify JWT access token nếu `server-service` tự kiểm tra auth.
- **`microservices/pkg/config`**: struct config chung như Postgres, Redis, JWT.
- **`microservices/pkg/db`**: tiện ích mở kết nối Postgres dùng chung.

Nếu cần Kafka publisher cho server domain events, chỉ tạo shared package mới hoặc local infra riêng sau khi xác nhận pattern chung với các service tiếp theo. Không tự cài dependency mới nếu chưa được yêu cầu.

## 2. Xây dựng thư mục `microservices/server-service`

Triển khai service quản lý danh mục máy chủ độc lập chạy trên một cổng riêng (ví dụ: `:8082`). Service này chịu trách nhiệm cho Server Inventory Cataloging: CRUD server, validate dữ liệu server, list/filter/sort/pagination, import/export XLSX và publish event thay đổi server cho `monitor-service` đồng bộ local state.

Các file/thư mục chính cần có:

- **`api/openapi.yaml`**: trích xuất phần spec liên quan đến server inventory: `GET /servers`, `POST /servers`, `PATCH /servers/{server_id}`, `DELETE /servers/{server_id}`, `POST /servers/import`, `GET /servers/export`.
- **`cmd/main.go`**: điểm khởi động ứng dụng. Khởi tạo config, Postgres, JWT middleware nếu có, Kafka publisher nếu có, repository, service, handler và router.
- **`internal/config/config.go`**: load cấu hình môi trường riêng cho `server-service`.
- **`internal/model/server.go`**: struct server map database Postgres và các request/response model liên quan đến server metadata.
- **`internal/domain/repo/serverRepoInterface.go`**: interface repository server.
- **`internal/domain/service/serverServiceInterface.go`**: interface logic nghiệp vụ server.
- **`internal/service/serverService.go`**: logic CRUD/list/import/export server.
- **`internal/infra/postgres/serverRepo.go`**: implementation repository bằng Postgres/GORM.
- **`internal/infra/file/deserialize/*`**: importer XLSX nếu import server vẫn dùng Excel.
- **`internal/infra/file/export/*`**: exporter XLSX nếu export server vẫn dùng Excel.
- **`internal/infra/kafka/*`**: publisher domain events server, chỉ khi phase này triển khai event publishing.
- **`internal/handler/serverHandler.go`**: handler cho CRUD/import/export server, bỏ logic report như `GenerateServerReport` và `GetReportFile`.

Không đưa các chức năng sau vào `server-service`:

- Heartbeat ingestion.
- Active ping scheduling.
- Uptime analytics/report generation.
- Report file download.
- Mail dispatch.
- Auth login/refresh/logout.

## 3. Cấu trúc service code mục tiêu

`server-service` cần tổ chức cùng style với `auth-service`: API/handler chỉ nhận request và map response, service giữ nghiệp vụ, domain khai báo interface, infra hiện thực kết nối ngoài, còn `microservices/pkg` chứa phần dùng chung giữa các service.

```text
microservices/
+-- pkg/
|   +-- apperr/          # error dùng chung
|   +-- auth/            # Role, Scope dùng trong authorization
|   +-- config/          # config structs dùng chung
|   +-- db/              # Postgres connection helper
|   +-- jwt/             # JWTProvider và claims
+-- server-service/
    +-- api/             # openapi.yaml và api.gen.go nếu đã được chuẩn bị
    +-- cmd/             # main.go bootstrap service
    +-- internal/
        +-- config/      # load env riêng cho server-service
        +-- domain/
        |   +-- repo/    # ServerRepoInterface
        |   +-- service/ # ServerServiceInterface
        +-- handler/     # OpenAPI/Gin handler implementation
        +-- infra/
        |   +-- postgres/# ServerRepo GORM implementation
        |   +-- file/    # XLSX import/export implementation
        |   +-- kafka/   # Server event publisher nếu triển khai phase này
        |   +-- runtime/ # App struct, DB/JWT/Kafka bootstrap nếu cần
        +-- model/       # Server và request/response model
        +-- service/     # ServerService nghiệp vụ
```

Luồng dependency cần giữ trong quá trình migration:

1. `handler.ServerHandler` phụ thuộc `domain/service.ServerServiceInterface`, không phụ thuộc trực tiếp infra.
2. `service.ServerService` phụ thuộc `domain/repo.ServerRepoInterface`, các importer/exporter interface nếu cần, và event publisher interface nếu phase này có publish Kafka.
3. `infra/postgres.ServerRepo` implement `domain/repo.ServerRepoInterface` bằng GORM.
4. `infra/file` giữ implementation import/export XLSX, không đặt parsing file trực tiếp trong handler.
5. `infra/kafka` giữ implementation publish `ServerCreated`, `ServerUpdated`, `ServerDeleted` nếu phase này bật domain events.
6. `cmd/main.go` là nơi wiring concrete dependencies: config/runtime -> DB -> repo -> file importer/exporter -> event publisher -> server service -> server handler -> router.
7. `internal/infra/runtime.App` chỉ giữ tài nguyên hạ tầng dùng lúc bootstrap; nghiệp vụ không đặt trong runtime.

Các struct chính cần giữ ổn định:

- `service.ServerService`: gồm `serverRepo`, importer/exporter nếu có, event publisher nếu có.
- `handler.ServerHandler`: gồm `serverService`.
- `runtime.App`: gồm `Config`, `DB`, `JWTProvider` nếu service tự verify token, và Kafka publisher/client nếu phase này triển khai event publishing.
- `model.Server`: giữ các field server metadata đang được Postgres và API dùng.

## 4. API contract cần tách

Giữ các endpoint server inventory:

- `GET /servers`
- `POST /servers`
- `PATCH /servers/{server_id}`
- `DELETE /servers/{server_id}`
- `POST /servers/import`
- `GET /servers/export`

Không đưa vào server-service:

- `POST /servers/report`
- `GET /report/{filename}`
- `/auth/*`
- `/heartbeat`

## 5. Domain events dự kiến

Khi server inventory thay đổi, `server-service` nên publish event để `monitor-service` đồng bộ danh sách server cần theo dõi:

- `ServerCreated`
- `ServerUpdated`
- `ServerDeleted`
- `ServersImported` hoặc nhiều event `ServerCreated`/`ServerUpdated` theo từng record import.

Payload tối thiểu:

```json
{
  "event_id": "uuid",
  "event_type": "ServerCreated",
  "server_id": "srv-001",
  "server_name": "web-01",
  "ipv4": "10.0.0.1",
  "occurred_at": "2026-06-22T00:00:00Z"
}
```

Phase đầu có thể để Kafka publisher là TODO nếu chưa có shared MQ package ổn định. Không tự thêm command/dependency để hoàn thiện Kafka nếu chưa được yêu cầu.

## 6. Auth và authorization

`server-service` cần bảo vệ endpoint bằng JWT access token và scope:

- `server:read` cho list/export.
- `server:create` cho create/import.
- `server:update` cho update.
- `server:delete` cho delete.

Có thể triển khai middleware local dựa trên `microservices/pkg/jwt` và `microservices/pkg/auth`, hoặc tạm đánh dấu TODO nếu phase này chỉ migrate business code.

## 7. Config dự kiến

Các biến môi trường chính:

- `APP_HOST`
- `APP_PORT`
- `APP_CORS_ORIGIN`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_DBNAME`
- `JWT_ACCESS_SECRET`
- `KAFKA_BROKER` nếu publish domain events.
- `KAFKA_SERVER_EVENT_TOPIC` nếu publish domain events.

## 8. Các bước thực thi (Execution Steps)

1. Tạo cấu trúc thư mục cho `microservices/server-service`.
2. Trích xuất OpenAPI contract chỉ gồm endpoint server inventory.
3. Copy/chuyển model server liên quan sang `internal/model` của service mới.
4. Copy/chuyển repo interface và Postgres implementation sang `internal/domain/repo` và `internal/infra/postgres`.
5. Copy/chuyển `serverService.go` sang `internal/service`, cập nhật import path sang `microservices/pkg` và package local.
6. Copy/chuyển handler server sang `internal/handler`, bỏ report/download logic khỏi handler.
7. Copy/chuyển XLSX import/export infra nếu service vẫn giữ import/export Excel.
8. Tạo `internal/config/config.go` để load app, Postgres, JWT và Kafka config nếu cần publish event.
9. Tạo `cmd/main.go` để wire config, DB, repository, service, handler và Gin router.
10. Kiểm tra lại dependency flow để đảm bảo chỉ tầng bootstrap tạo concrete infra, còn handler/service vẫn nhận interface theo cấu trúc ở mục 3.
11. Ghi chú các manual follow-up command cho người dùng nếu cần, nhưng không tự chạy.

## 9. Manual follow-up sau migration

Sau khi code được migrate, người dùng có thể tự chạy nếu muốn:

- `go mod tidy` trong `microservices/server-service`.
- API code generation nếu dự án quyết định dùng generated server từ OpenAPI.
- Build/test thủ công.
- Docker compose/swarm update thủ công.

Agent không tự chạy các command trên trong migration.
