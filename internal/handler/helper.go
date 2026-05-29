package handler

import (
	"github.com/LeHuuHai/server-management/api"
)

func BadRequest(err error) api.BadRequestJSONResponse {
	msg := err.Error()
	code := "400"
	return api.BadRequestJSONResponse{Message: &msg, Code: &code}
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
