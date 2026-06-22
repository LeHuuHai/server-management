package handler

import "github.com/LeHuuHai/server-management/microservices/auth-service/api"

func Unauthorized(err error) api.UnauthorizedJSONResponse {
	msg := err.Error()
	code := "UNAUTHORIZED"
	return api.UnauthorizedJSONResponse{Message: &msg, Code: &code}
}

func InternalError(err error) api.InternalErrorJSONResponse {
	msg := err.Error()
	code := "INTERNAL_ERROR"
	return api.InternalErrorJSONResponse{Message: &msg, Code: &code}
}
