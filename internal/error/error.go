package apperr

import "errors"

var (
	ErrInvalidIP         = errors.New("invalid ipv4")
	ErrInvalidSort       = errors.New("invalid sort field or order")
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrDuplicateServer   = errors.New("duplicate server id or server name")
	ErrRecordNotFound    = errors.New("record not found")
	ErrInvalidImportData = errors.New("file have invalid data or format")
)
