module github.com/LeHuuHai/server-management/microservices/auth-service

go 1.25.3

require (
	github.com/LeHuuHai/server-management/microservices/pkg v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.19.0
	golang.org/x/crypto v0.48.0
	google.golang.org/grpc v1.81.1
	gorm.io/gorm v1.25.10
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
)

replace github.com/LeHuuHai/server-management/microservices/pkg => ../pkg
