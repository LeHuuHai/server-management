# Kế hoạch di chuyển Auth Service (auth-service)

Kế hoạch này chi tiết các bước bóc tách thành phần xác thực (Auth) từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/auth-service` và thư viện dùng chung `microservices/pkg` dựa trên giao tiếp gRPC.

> **QUAN TRỌNG:** Theo kiến trúc mới, `auth-service` sẽ chỉ phục vụ qua gRPC. Phần API Gateway (chứa HTTP REST handlers) sẽ được di chuyển ở một kế hoạch khác. Trong phạm vi tài liệu này, chúng ta chỉ tập trung thiết lập server RPC cho Auth.

## 1. Tách các Shared Packages và gRPC Contracts vào `microservices/pkg`

Các file chung và hợp đồng giao tiếp (Contracts) sẽ được đặt ở đây:

- **`microservices/pkg/apperr/error.go`**: Chứa custom error. Nguyên gốc từ `internal/error/error.go`.
- **`microservices/pkg/auth/role.go`**: Quản lý phân quyền Scope theo Role.
- **`microservices/pkg/auth/scope.go`**: Định nghĩa danh sách các Scope.
- **`microservices/pkg/jwt/jwtProvider.go`**: Logic JWT token dùng để verify token.
- **`microservices/pkg/config/config.go`**: Struct định nghĩa chung về config.
- **`microservices/pkg/db/db.go`**: Tiện ích mở kết nối tới database Postgres.
- **`microservices/pkg/cache/cache.go`**: Tiện ích mở kết nối tới Redis.
- **`microservices/pkg/pb/auth/`**: Nơi chứa file `auth.proto` định nghĩa gRPC services cho Auth (`Login`, `RefreshToken`, `Logout`).

## 2. Xây dựng thư mục `microservices/auth-service` (gRPC Server)

Triển khai service xác thực độc lập chạy trên một cổng gRPC (ví dụ: `:50051`). Service này sẽ **không chứa HTTP Gin Router**:

- **`cmd/main.go`**: Điểm khởi động. Khởi tạo DB, Redis, cấu hình JWT, repositories, services, sau đó mở kết nối `net.Listen` và chạy `grpc.NewServer()`.
- **`internal/config/config.go`**: Load cấu hình môi trường riêng cho `auth-service`.
- **`internal/model/*`**: Các models nội bộ (Account, User, LoginResult).
- **`internal/domain/cache/tokenBlocklist.go`**: Interface blocklist cache.
- **`internal/domain/repo/accountRepoInterface.go`**: Interface repo tài khoản.
- **`internal/domain/service/authServiceInterface.go`**: Interface logic nghiệp vụ.
- **`internal/infra/postgres/accountRepo.go`**: Implement thao tác Postgres.
- **`internal/infra/redis/tokenBlocklist.go`**: Implement thao tác Redis.
- **`internal/service/authService.go`**: Logic lõi cho Login, Logout, HashPassword, RefreshAccessToken.
- **`internal/rpc/auth_server.go`**: Implementation của gRPC Server interface sinh ra từ proto, nhận request gRPC, chuyển đổi struct và gọi xuống tầng `service.AuthService`. (Chức năng này thay thế cho `handler/authHandler.go` cũ).

## 3. Luồng dependency và Cấu trúc tổng thể

```text
microservices/
├── pkg/
│   ├── apperr/
│   ├── auth/
│   ├── cache/
│   ├── config/
│   ├── db/
│   ├── jwt/
│   └── pb/auth/       # Chỉ chứa file auth.proto
└── auth-service/      # Chỉ mở gRPC
    ├── cmd/           # main.go (Khởi động gRPC server)
    └── internal/
        ├── config/
        ├── domain/
        ├── infra/
        │   └── service/ # Logic nghiệp vụ AuthService
        ├── model/
        └── rpc/       # gRPC Server handler
```

## 4. Các bước thực thi (Execution Steps)

**Lưu ý:** Agent sẽ KHÔNG chạy lệnh cài đặt dependency, KHÔNG chạy lệnh sinh code (protoc), và KHÔNG chạy `go mod tidy`. Người dùng sẽ tự thực hiện các bước này.

1. **Khai báo Proto**: Tạo thư mục `microservices/pkg/pb/auth` và viết file `auth.proto` mô tả cấu trúc request/response. (Chỉ tạo file proto, để lại phần gen code cho người dùng).
2. **Di chuyển Shared Packages**: Chép các package dùng chung (error, auth, jwt, config, db, cache) từ khối monolith vào `microservices/pkg`.
3. **Xây dựng `auth-service`**: Tạo cấu trúc thư mục `microservices/auth-service`. Di chuyển các thành phần Domain, Infra, Model, Service.
4. **Chuẩn bị file gRPC Handler**: Viết code khung cho `internal/rpc/auth_server.go` trong `auth-service` để map các hàm RPC vào `AuthService` dựa trên cấu trúc sinh ra dự kiến từ `auth.proto`.
5. **Khởi động gRPC**: Cập nhật `cmd/main.go` cho `auth-service` để khởi tạo port RPC và Listen.
6. Bàn giao lại để người dùng tự chạy `protoc`, cài đặt gRPC packages và `go mod tidy`.
