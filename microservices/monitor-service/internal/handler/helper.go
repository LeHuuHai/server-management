package handler

import (
	"github.com/LeHuuHai/server-management/microservices/monitor-service/api"
)

func BadRequest(err error) api.BadRequestJSONResponse {
	msg := err.Error()
	code := "400"
	return api.BadRequestJSONResponse{Message: &msg, Code: &code}
}

func InternalError(err error) api.InternalErrorJSONResponse {
	msg := err.Error()
	code := "500"
	return api.InternalErrorJSONResponse{Message: &msg, Code: &code}
}
