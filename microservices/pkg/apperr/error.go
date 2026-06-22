package apperr

import "errors"

var (
	ErrInvalidIP            = errors.New("invalid ipv4")
	ErrInvalidSort          = errors.New("invalid sort field or order")
	ErrInvalidPagination    = errors.New("invalid pagination")
	ErrDuplicateServer      = errors.New("duplicate server id or server name")
	ErrRecordNotFound       = errors.New("record not found")
	ErrInvalidImportData    = errors.New("file have invalid data or format")
	ErrInvalidTimeRange     = errors.New("invalid time range")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrConnectPostgres      = errors.New("connect postgres failed")
	ErrConnectElasticsearch = errors.New("connect elasticsearch failed")
	ErrConnectKafka         = errors.New("connect kafka failed")
	ErrConnectRedis         = errors.New("connect redis failed")
	ErrAppBuild             = errors.New("build app failed")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDeniedAccess       = errors.New("access denied")
	ErrSignToken          = errors.New("sign token failed")
	ErrExpiredToken       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrRevokedToken       = errors.New("revoked token")
)
