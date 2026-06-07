package xlsxexport_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	xlsxexport "github.com/LeHuuHai/server-management/internal/infra/file/export"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportServerXLSX_Success(t *testing.T) {
	now := time.Now()
	servers := []model.Server{
		{ServerID: "s1", ServerName: "Server1", IPv4: "1.2.3.4", Status: model.StatusOnline, CreatedAt: now},
		{ServerID: "s2", ServerName: "Server2", IPv4: "5.6.7.8", Status: model.StatusOffline, CreatedAt: now},
	}

	exporter := xlsxexport.NewServerXLSXExporter()
	buf := &bytes.Buffer{}
	err := exporter.Export(context.Background(), buf, servers)

	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}

func TestExportServerXLSX_EmptyData(t *testing.T) {
	exporter := xlsxexport.NewServerXLSXExporter()
	buf := &bytes.Buffer{}
	err := exporter.Export(context.Background(), buf, []model.Server{})

	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}
