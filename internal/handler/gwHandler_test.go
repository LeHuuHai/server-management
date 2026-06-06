package handler_test

import (
	"context"
	"errors"
	"testing"

	gwapi "github.com/LeHuuHai/server-management/api/gw"
	"github.com/LeHuuHai/server-management/internal/handler"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// SendHeartbeat
// ---------------------------------------------------------------------------

func TestGwHandler_SendHeartbeat_Success(t *testing.T) {
	mockGW := mocks.NewMockGWServiceInterface(t)
	mockGW.EXPECT().PublishHeartbeat(mock.Anything, model.Heartbeat{
		ServerID: "srv-01",
	})

	handler := handler.NewGwHandler(mockGW)
	resp, err := handler.SendHeartbeat(context.Background(), gwapi.SendHeartbeatRequestObject{
		Body: &gwapi.SendHeartbeatJSONRequestBody{ServerId: "srv-01"},
	})

	assert.NoError(t, err)
	_, ok := resp.(gwapi.SendHeartbeat202Response)
	assert.True(t, ok)
	mockGW.AssertExpectations(t)
}

func TestGwHandler_SendHeartbeat_PublishError_Returns500(t *testing.T) {
	mockGW := mocks.NewMockGWServiceInterface(t)
	mockGW.EXPECT().PublishHeartbeat(mock.Anything, mock.Anything).Return(errors.New("kfk error"))

	handler := handler.NewGwHandler(mockGW)
	resp, err := handler.SendHeartbeat(context.Background(), gwapi.SendHeartbeatRequestObject{
		Body: &gwapi.SendHeartbeatJSONRequestBody{ServerId: "any"},
	})

	assert.NoError(t, err)
	res, ok := resp.(gwapi.SendHeartbeat500JSONResponse)
	assert.True(t, ok)
	assert.Equal(t, "kfk error", *res.Message)
	mockGW.AssertExpectations(t)
}

func TestGwHandler_SendHeartbeat_UnexpectedError_Returns500(t *testing.T) {
	mockGW := mocks.NewMockGWServiceInterface(t)
	mockGW.EXPECT().PublishHeartbeat(mock.Anything, mock.Anything).Return(errors.New("other error"))

	handler := handler.NewGwHandler(mockGW)
	resp, err := handler.SendHeartbeat(context.Background(), gwapi.SendHeartbeatRequestObject{
		Body: &gwapi.SendHeartbeatJSONRequestBody{ServerId: "any"},
	})

	assert.NoError(t, err)
	_, ok := resp.(gwapi.SendHeartbeat500JSONResponse)
	assert.True(t, ok)
	mockGW.AssertExpectations(t)
}
