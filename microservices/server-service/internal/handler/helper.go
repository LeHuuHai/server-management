package handler

import (
	"github.com/LeHuuHai/server-management/microservices/server-service/api"
)

func BadRequest(err error) api.BadRequestJSONResponse {
	msg := err.Error()
	code := "400"
	return api.BadRequestJSONResponse{Message: &msg, Code: &code}
}

func Unauthorized(err error) api.UnauthorizedJSONResponse {
	msg := err.Error()
	code := "401"
	return api.UnauthorizedJSONResponse{Message: &msg, Code: &code}
}

func Forbidden(err error) api.ForbiddenJSONResponse {
	msg := err.Error()
	code := "403"
	return api.ForbiddenJSONResponse{Message: &msg, Code: &code}
}

func NotFound(err error) api.NotFoundJSONResponse {
	msg := err.Error()
	code := "404"
	return api.NotFoundJSONResponse{Message: &msg, Code: &code}
}

func Conflict(err error) api.ConflictJSONResponse {
	msg := err.Error()
	code := "409"
	return api.ConflictJSONResponse{Message: &msg, Code: &code}
}

func InternalError(err error) api.InternalErrorJSONResponse {
	msg := err.Error()
	code := "500"
	return api.InternalErrorJSONResponse{Message: &msg, Code: &code}
}
