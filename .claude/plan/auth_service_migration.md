# Kế hoạch di chuyển Auth Service (auth-service)

Kế hoạch này chi tiết các bước bóc tách thành phần xác thực (Auth) từ Monolith cũ sang Bounded Context mới dưới thư mục `microservices/auth-service` và thư viện dùng chung `microservices/pkg` dựa trên phân tích cấu trúc dự án.

## 1. Tách các Shared Packages vào `microservices/pkg`

Các file chung sẽ được tách và di chuyển để các service khác có thể sử dụng lại:

- **`microservices/pkg/apperr/error.go`**: Chứa custom error. Nguyên gốc từ `internal/error/error.go`.
- **`microservices/pkg/auth/role.go`**: Quản lý phân quyền Scope theo Role. Nguyên gốc từ `internal/domain/auth/role.go`.
- **`microservices/pkg/auth/scope.go`**: Định nghĩa danh sách các Scope. Nguyên gốc từ `internal/domain/auth/scope.go`.
- **`microservices/pkg/jwt/jwtProvider.go`**: Logic JWT token dùng để verify token ở các service khác. Nguyên gốc từ `internal/infra/jwt/jwtProvider.go`.
- **`microservices/pkg/config/config.go`**: Struct định nghĩa chung về config của Postgres, Redis, JWT.
- **`microservices/pkg/db/db.go`**: Tiện ích mở kết nối tới database Postgres dùng chung.
- **`microservices/pkg/cache/cache.go`**: Tiện ích mở kết nối tới Redis dùng chung.

## 2. Xây dựng thư mục `microservices/auth-service`

Triển khai service xác thực độc lập chạy trên một cổng riêng (ví dụ: `:8081`). Chúng ta sẽ tái sử dụng spec OpenAPI và implementation handler hiện có:

- **`api/openapi.yaml`**: Trích xuất phần spec liên quan đến auth (`/auth/login`, `/auth/refresh`, `/auth/logout`) để dùng cho việc gen code sau này.
- **`cmd/main.go`**: Điểm khởi động ứng dụng. Khởi tạo DB, Redis, cấu hình JWT, repositories, services và router.
- **`internal/config/config.go`**: Load cấu hình môi trường.
- **`internal/model/account.go`**: Struct `Account` map database Postgres.
- **`internal/model/user.go`**: Struct `User`.
- **`internal/model/loginResult.go`**: Kết quả trả về sau login.
- **`internal/domain/cache/tokenBlocklist.go`**: Interface blocklist cache.
- **`internal/domain/repo/accountRepoInterface.go`**: Interface repo tài khoản.
- **`internal/domain/service/authServiceInterface.go`**: Interface logic nghiệp vụ xác thực.
- **`internal/infra/postgres/accountRepo.go`**: Implement thao tác Postgres.
- **`internal/infra/redis/tokenBlocklist.go`**: Implement thao tác Redis.
- **`internal/service/authService.go`**: Logic Login, Logout, HashPassword, RefreshAccessToken.
- **`internal/handler/authHandler.go`**: Tái sử dụng implementation hiện tại của auth handler để gắn vào code gen.

## 3. Cấu trúc service code hiện tại

Auth service đang được tổ chức theo hướng tách lớp rõ ràng: API/handler chỉ nhận request và map response, service giữ nghiệp vụ, domain khai báo interface, infra hiện thực kết nối ngoài, còn `microservices/pkg` chứa phần dùng chung giữa các service.

```text
microservices/
├── pkg/
│   ├── apperr/          # error dùng chung
│   ├── auth/            # Role, Scope dùng trong claims và authorization
│   ├── cache/           # Redis connection helper
│   ├── config/          # PostgresConfig, RedisConfig, JWTConfig
│   ├── db/              # Postgres connection helper
│   └── jwt/             # JWTProvider và claims
└── auth-service/
    ├── api/             # openapi.yaml và api.gen.go
    ├── cmd/             # main.go bootstrap service
    └── internal/
        ├── config/      # load env riêng cho auth-service
        ├── domain/
        │   ├── cache/   # TokenBlocklist interface
        │   ├── repo/    # AccountRepoInterface
        │   └── service/ # AuthServiceInterface
        ├── handler/     # OpenAPI strict handler implementation
        ├── infra/
        │   ├── postgres/# AccountRepo GORM implementation
        │   ├── redis/   # TokenBlocklistRedis implementation
        │   └── runtime/ # App struct, DB/Redis/JWT bootstrap
        ├── model/       # Account, User, LoginResult
        └── service/     # AuthService nghiệp vụ
```

Luồng dependency cần giữ trong quá trình migration:

1. `handler.AuthHandler` phụ thuộc `domain/service.AuthServiceInterface`, không phụ thuộc trực tiếp infra.
2. `service.AuthService` phụ thuộc `pkg/jwt.JWTProvider`, `domain/cache.TokenBlocklist`, `domain/repo.AccountRepoInterface`.
3. `infra/postgres.AccountRepo` implement `domain/repo.AccountRepoInterface` bằng GORM.
4. `infra/redis.TokenBlocklistRedis` implement `domain/cache.TokenBlocklist` bằng go-redis.
5. `cmd/main.go` là nơi wiring concrete dependencies: config/runtime -> JWT provider -> Redis blocklist -> Postgres repo -> auth service -> auth handler -> OpenAPI strict handler -> Gin router.
6. `internal/infra/runtime.App` chỉ giữ tài nguyên hạ tầng dùng lúc bootstrap (`Config`, `JWTProvider`, `DB`, `RdbClient`); nghiệp vụ không đặt trong runtime.

Các struct chính cần giữ ổn định:

- `service.AuthService`: gồm `jwtProvider`, `blocklist`, `accountRepo`.
- `handler.AuthHandler`: gồm `authService`.
- `runtime.App`: gồm `Config`, `JWTProvider`, `DB`, `RdbClient`.
- `model.Account`: implement `pkg/jwt.AccountClaimsSource` qua `GetUserID()` và `GetRole()`.

## 4. Các bước thực thi (Execution Steps)

1. Tạo thư mục `microservices/pkg` và sao chép các file dùng chung (`error.go`, `role.go`, `scope.go`, `jwtProvider.go`, DB utils).
2. Tạo cấu trúc thư mục cho `microservices/auth-service`.
3. Trích xuất các endpoint liên quan đến auth từ `api/openapi.yaml` cũ sang `microservices/auth-service/api/openapi.yaml`.
4. Sao chép các models nội bộ (`account.go`, `user.go`, `loginResult.go`) và các file interface/implementation (domain, infra, service) sang `auth-service`.
5. Sao chép `internal/handler/authHandler.go` hiện tại sang `auth-service` (để dùng với server/router được gen ra từ spec).
6. Khởi tạo `cmd/main.go` để chạy service độc lập.
7. Cập nhật import paths cho phù hợp với cấu trúc module mới. (User sẽ tự chạy `go mod tidy` sau cùng).
8. Kiểm tra lại wiring trong `cmd/main.go` để đảm bảo chỉ tầng bootstrap tạo concrete infra, còn handler/service vẫn nhận interface theo cấu trúc ở mục 3.
