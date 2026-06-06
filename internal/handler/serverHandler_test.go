package handler_test

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/LeHuuHai/server-management/api"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/handler"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/mocks"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GetListServers
// ---------------------------------------------------------------------------

func TestServerHandler_GetListServers_Success(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	now := time.Now()
	filter := model.ListServerFilter{From: 0, To: 10, SortField: model.SortByName, Desc: false}
	mockSvc.EXPECT().ListServer(mock.Anything, filter).Return(&model.ListServerResult{
		Servers: []model.Server{
			{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4", Status: model.StatusOnline, CreatedAt: now},
		},
		Total: 1,
	}, nil)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	sortField := string(model.SortByName)
	resp, err := handler.GetListServers(context.Background(), api.GetListServersRequestObject{
		Params: api.GetListServersParams{From: 0, To: 10, SortField: sortField},
	})

	assert.NoError(t, err)
	res, ok := resp.(api.GetListServers200JSONResponse)
	assert.True(t, ok)
	assert.Equal(t, 1, *res.Total)
	assert.Len(t, *res.Items, 1)
	assert.Equal(t, "s1", (*res.Items)[0].ServerId)
}

func TestServerHandler_GetListServers_InvalidSort_Returns400(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(nil, apperr.ErrInvalidSort)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.GetListServers(context.Background(), api.GetListServersRequestObject{
		Params: api.GetListServersParams{From: 0, To: 10, SortField: "invalid"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.GetListServers400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_GetListServers_InvalidPagination_Returns400(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(nil, apperr.ErrInvalidPagination)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.GetListServers(context.Background(), api.GetListServersRequestObject{
		Params: api.GetListServersParams{From: -1, To: 0, SortField: string(model.SortByName)},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.GetListServers400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_GetListServers_ServiceError_Returns500(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(nil, errors.New("some error, such as db error"))

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.GetListServers(context.Background(), api.GetListServersRequestObject{
		Params: api.GetListServersParams{From: 0, To: 10, SortField: string(model.SortByName)},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.GetListServers500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// CreateServer
// ---------------------------------------------------------------------------

func TestServerHandler_CreateServer_Success(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	now := time.Now()
	mockSvc.EXPECT().CreateServer(mock.Anything, mock.MatchedBy(func(s *model.Server) bool {
		return s.ServerID == "srv-01" && s.IPv4 == "192.168.1.1"
	})).Return(&model.Server{
		ServerID:   "srv-01",
		ServerName: "My Server",
		IPv4:       "192.168.1.1",
		Status:     model.StatusUnknown,
		CreatedAt:  now,
	}, nil)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.CreateServer(context.Background(), api.CreateServerRequestObject{
		Body: &api.CreateServerJSONRequestBody{ServerId: "srv-01", ServerName: "My Server", Ipv4: "192.168.1.1"},
	})

	assert.NoError(t, err)
	res, ok := resp.(api.CreateServer201JSONResponse)
	assert.True(t, ok)
	assert.Equal(t, "srv-01", res.ServerId)
}

func TestServerHandler_CreateServer_InvalidIP_Returns400(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().CreateServer(mock.Anything, mock.Anything).Return(nil, apperr.ErrInvalidIP)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.CreateServer(context.Background(), api.CreateServerRequestObject{
		Body: &api.CreateServerJSONRequestBody{ServerId: "s1", ServerName: "S1", Ipv4: "not-an-ip"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.CreateServer400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_CreateServer_Duplicate_Returns409(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().CreateServer(mock.Anything, mock.Anything).Return(nil, apperr.ErrDuplicateServer)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.CreateServer(context.Background(), api.CreateServerRequestObject{
		Body: &api.CreateServerJSONRequestBody{ServerId: "s1", ServerName: "S1", Ipv4: "1.2.3.4"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.CreateServer409JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_CreateServer_ServiceError_Returns500(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().CreateServer(mock.Anything, mock.Anything).Return(nil, errors.New("some err, such as db error"))

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.CreateServer(context.Background(), api.CreateServerRequestObject{
		Body: &api.CreateServerJSONRequestBody{ServerId: "s1", ServerName: "S1", Ipv4: "1.2.3.4"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.CreateServer500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// DeleteServer
// ---------------------------------------------------------------------------

func TestServerHandler_DeleteServer_Success(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().DeleteServer(mock.Anything, "srv-01").Return(nil)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.DeleteServer(context.Background(), api.DeleteServerRequestObject{ServerId: "srv-01"})

	assert.NoError(t, err)
	_, ok := resp.(api.DeleteServer204Response)
	assert.True(t, ok)
}

func TestServerHandler_DeleteServer_NotFound_Returns404(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().DeleteServer(mock.Anything, "ghost").Return(apperr.ErrRecordNotFound)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.DeleteServer(context.Background(), api.DeleteServerRequestObject{ServerId: "ghost"})

	assert.NoError(t, err)
	_, ok := resp.(api.DeleteServer404JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_DeleteServer_ServiceError_Returns500(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().DeleteServer(mock.Anything, "srv-01").Return(errors.New("some err, such as db error"))

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.DeleteServer(context.Background(), api.DeleteServerRequestObject{ServerId: "srv-01"})

	assert.NoError(t, err)
	_, ok := resp.(api.DeleteServer500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// UpdateServer
// ---------------------------------------------------------------------------

func TestServerHandler_UpdateServer_Success(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	name, ip := "Updated", "10.0.0.1"
	now := time.Now()
	mockSvc.EXPECT().UpdateServer(mock.Anything, mock.MatchedBy(func(s *model.Server) bool {
		return s.ServerID == "srv-01"
	})).Return(&model.Server{
		ServerID:   "srv-01",
		ServerName: "Updated",
		IPv4:       "10.0.0.1",
		Status:     model.StatusOnline,
		CreatedAt:  now,
	}, nil)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.UpdateServer(context.Background(), api.UpdateServerRequestObject{
		ServerId: "srv-01",
		Body:     &api.UpdateServerJSONRequestBody{ServerName: &name, Ipv4: &ip},
	})

	assert.NoError(t, err)
	res, ok := resp.(api.UpdateServer200JSONResponse)
	assert.True(t, ok)
	assert.Equal(t, "srv-01", res.ServerId)
	assert.Equal(t, "Updated", res.ServerName)
}

func TestServerHandler_UpdateServer_NotFound_Returns404(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	name, ip := "X", "1.2.3.4"
	mockSvc.EXPECT().UpdateServer(mock.Anything, mock.Anything).Return(nil, apperr.ErrRecordNotFound)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.UpdateServer(context.Background(), api.UpdateServerRequestObject{
		ServerId: "ghost",
		Body:     &api.UpdateServerJSONRequestBody{ServerName: &name, Ipv4: &ip},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.UpdateServer404JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_UpdateServer_ServiceError_Returns500(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	name, ip := "X", "1.2.3.4"
	mockSvc.EXPECT().UpdateServer(mock.Anything, mock.Anything).Return(nil, errors.New("some err, such as db error"))

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.UpdateServer(context.Background(), api.UpdateServerRequestObject{
		ServerId: "srv-01",
		Body:     &api.UpdateServerJSONRequestBody{ServerName: &name, Ipv4: &ip},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.UpdateServer500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// GenerateServerReport
// ---------------------------------------------------------------------------

func TestServerHandler_GenerateServerReport_Success(t *testing.T) {
	mockReport := mocks.NewMockReportServiceInterface(t)
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	mockReport.EXPECT().ReportServer(mock.Anything, model.GenServerReportRequest{
		From:      from,
		To:        to,
		Receivers: []string{"admin@example.com"},
	}).Return(nil)

	handler := handler.NewServerHandler(nil, mockReport, nil, nil)
	resp, err := handler.GenerateServerReport(context.Background(), api.GenerateServerReportRequestObject{
		Body: &api.GenerateServerReportJSONRequestBody{
			From:      from,
			To:        to,
			Receivers: []openapi_types.Email{"admin@example.com"},
		},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.GenerateServerReport202Response)
	assert.True(t, ok)
}

func TestServerHandler_GenerateServerReport_InvalidTimeRange_Returns400(t *testing.T) {
	mockReport := mocks.NewMockReportServiceInterface(t)
	mockReport.EXPECT().ReportServer(mock.Anything, mock.Anything).Return(apperr.ErrInvalidTimeRange)

	handler := handler.NewServerHandler(nil, mockReport, nil, nil)
	resp, err := handler.GenerateServerReport(context.Background(), api.GenerateServerReportRequestObject{
		Body: &api.GenerateServerReportJSONRequestBody{
			From:      time.Now(),
			To:        time.Now().Add(-time.Hour),
			Receivers: []openapi_types.Email{"admin@example.com"},
		},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.GenerateServerReport400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_GenerateServerReport_InvalidEmail_Returns400(t *testing.T) {
	mockReport := mocks.NewMockReportServiceInterface(t)
	mockReport.EXPECT().ReportServer(mock.Anything, mock.Anything).Return(apperr.ErrInvalidEmail)

	handler := handler.NewServerHandler(nil, mockReport, nil, nil)
	resp, err := handler.GenerateServerReport(context.Background(), api.GenerateServerReportRequestObject{
		Body: &api.GenerateServerReportJSONRequestBody{
			From:      time.Now(),
			To:        time.Now().Add(time.Hour),
			Receivers: []openapi_types.Email{"not-an-email"},
		},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.GenerateServerReport400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_GenerateServerReport_ServiceError_Returns500(t *testing.T) {
	mockReport := mocks.NewMockReportServiceInterface(t)
	mockReport.EXPECT().ReportServer(mock.Anything, mock.Anything).Return(errors.New("some err, such as es down"))

	handler := handler.NewServerHandler(nil, mockReport, nil, nil)
	resp, err := handler.GenerateServerReport(context.Background(), api.GenerateServerReportRequestObject{
		Body: &api.GenerateServerReportJSONRequestBody{
			From:      time.Now(),
			To:        time.Now().Add(time.Hour),
			Receivers: []openapi_types.Email{"admin@example.com"},
		},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.GenerateServerReport500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// ExportServers
// ---------------------------------------------------------------------------

func TestServerHandler_ExportServers_Success(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockExp := mocks.NewMockServerExporter(t)

	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(&model.ListServerResult{
		Servers: []model.Server{{ServerID: "s1", ServerName: "S1", IPv4: "1.1.1.1"}},
		Total:   1,
	}, nil)
	mockExp.EXPECT().Export(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockExp.EXPECT().FileType().Return("xlsx")

	handler := handler.NewServerHandler(mockSvc, nil, mockExp, nil)
	resp, err := handler.ExportServers(context.Background(), api.ExportServersRequestObject{
		Params: api.ExportServersParams{From: 0, To: 10, SortField: string(model.SortByName)},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ExportServers200ApplicationvndOpenxmlformatsOfficedocumentSpreadsheetmlSheetResponse)
	assert.True(t, ok)
}

func TestServerHandler_ExportServers_ListError_Returns500(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.ExportServers(context.Background(), api.ExportServersRequestObject{
		Params: api.ExportServersParams{From: 0, To: 10, SortField: string(model.SortByName)},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ExportServers500JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_ExportServers_ExportError_Returns500(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockExp := mocks.NewMockServerExporter(t)

	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(&model.ListServerResult{
		Servers: []model.Server{{ServerID: "s1"}},
		Total:   1,
	}, nil)
	mockExp.EXPECT().Export(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("write error"))

	handler := handler.NewServerHandler(mockSvc, nil, mockExp, nil)
	resp, err := handler.ExportServers(context.Background(), api.ExportServersRequestObject{
		Params: api.ExportServersParams{From: 0, To: 10, SortField: string(model.SortByName)},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ExportServers500JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_ExportServers_InvalidSort_Returns400(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(nil, apperr.ErrInvalidSort)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.ExportServers(context.Background(), api.ExportServersRequestObject{
		Params: api.ExportServersParams{From: 0, To: 10, SortField: "invalid"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ExportServers400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_ExportServers_InvalidPagination_Returns400(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockSvc.EXPECT().ListServer(mock.Anything, mock.Anything).Return(nil, apperr.ErrInvalidPagination)

	handler := handler.NewServerHandler(mockSvc, nil, nil, nil)
	resp, err := handler.ExportServers(context.Background(), api.ExportServersRequestObject{
		Params: api.ExportServersParams{From: -1, To: 0, SortField: string(model.SortByName)},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ExportServers400JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// ImportServer
// ---------------------------------------------------------------------------

func TestServerHandler_ImportServer_Success(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockDeserializer := mocks.NewMockServerDeserializer(t)

	servers := []model.ServerImport{
		{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4"},
		{ServerID: "s2", ServerName: "Server2", IPv4: "5.6.7.8"},
	}
	mockDeserializer.EXPECT().Deserialize(mock.Anything, mock.Anything).Return(servers, nil)
	mockSvc.EXPECT().ImportServer(mock.Anything, servers).Return(&model.CreateBatchServerResult{
		Success:    []string{"s1", "s2"},
		Failed:     []string{},
		SuccessCnt: 2,
		FailedCnt:  0,
	}, nil)

	handler := handler.NewServerHandler(mockSvc, nil, nil, mockDeserializer)
	body, writer := newMultipartBody(t, "servers.xlsx", []byte("fake-xlsx-content"))
	resp, err := handler.ImportServer(context.Background(), api.ImportServerRequestObject{
		Body: body,
	})

	_ = writer
	assert.NoError(t, err)
	res, ok := resp.(api.ImportServer200JSONResponse)
	assert.True(t, ok)
	assert.Equal(t, 2, res.TotalSuccess)
	assert.Equal(t, 0, res.TotalFailed)
}

func TestServerHandler_ImportServer_NoFile_Returns400(t *testing.T) {
	handler := handler.NewServerHandler(nil, nil, nil, nil)

	// body rỗng, NextPart() sẽ trả về lỗi
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.Close() // đóng ngay không có part nào
	emptyBody := multipart.NewReader(&buf, writer.Boundary())

	resp, err := handler.ImportServer(context.Background(), api.ImportServerRequestObject{
		Body: emptyBody,
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ImportServer400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_ImportServer_DeserializeInvalidData_Returns400(t *testing.T) {
	mockDeserializer := mocks.NewMockServerDeserializer(t)
	mockDeserializer.EXPECT().Deserialize(mock.Anything, mock.Anything).Return(nil, apperr.ErrInvalidImportData)

	handler := handler.NewServerHandler(nil, nil, nil, mockDeserializer)
	body, _ := newMultipartBody(t, "servers.xlsx", []byte("bad-content"))
	resp, err := handler.ImportServer(context.Background(), api.ImportServerRequestObject{
		Body: body,
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ImportServer400JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_ImportServer_DeserializeError_Returns500(t *testing.T) {
	mockDeserializer := mocks.NewMockServerDeserializer(t)
	mockDeserializer.EXPECT().Deserialize(mock.Anything, mock.Anything).Return(nil, errors.New("io error"))

	handler := handler.NewServerHandler(nil, nil, nil, mockDeserializer)
	body, _ := newMultipartBody(t, "servers.xlsx", []byte("content"))
	resp, err := handler.ImportServer(context.Background(), api.ImportServerRequestObject{
		Body: body,
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ImportServer500JSONResponse)
	assert.True(t, ok)
}

func TestServerHandler_ImportServer_ServiceError_Returns500(t *testing.T) {
	mockSvc := mocks.NewMockServerServiceInterface(t)
	mockDeserializer := mocks.NewMockServerDeserializer(t)

	servers := []model.ServerImport{{ServerID: "s1", ServerName: "S1", IPv4: "1.2.3.4"}}
	mockDeserializer.EXPECT().Deserialize(mock.Anything, mock.Anything).Return(servers, nil)
	mockSvc.EXPECT().ImportServer(mock.Anything, servers).Return(nil, errors.New("db error"))

	handler := handler.NewServerHandler(mockSvc, nil, nil, mockDeserializer)
	body, _ := newMultipartBody(t, "servers.xlsx", []byte("content"))
	resp, err := handler.ImportServer(context.Background(), api.ImportServerRequestObject{
		Body: body,
	})

	assert.NoError(t, err)
	_, ok := resp.(api.ImportServer500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// helper: tạo multipart body cho ImportServer
// ---------------------------------------------------------------------------

func newMultipartBody(t *testing.T, filename string, content []byte) (*multipart.Reader, *multipart.Writer) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	writer.Close()
	reader := multipart.NewReader(&buf, writer.Boundary())
	return reader, writer
}
