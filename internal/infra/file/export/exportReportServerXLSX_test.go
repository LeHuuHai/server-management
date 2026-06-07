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

func TestReportServerXLSXExporter_FileType(t *testing.T) {
	exporter := xlsxexport.NewReportServerXLSXExporter()
	assert.Equal(t, "xlsx", exporter.FileType())
}

func TestReportServerXLSXExporter_ContentType(t *testing.T) {
	exporter := xlsxexport.NewReportServerXLSXExporter()
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", exporter.ContentType())
}

func TestReportServerXLSXExporter_Success(t *testing.T) {
	now := time.Now()
	data := []model.ServerUptimeAgg{
		{ServerID: "s1", UptimeRatio: 0.99, StartPingAt: now, LastPingAt: now, DocCount: 100},
		{ServerID: "s2", UptimeRatio: 0.75, StartPingAt: now, LastPingAt: now, DocCount: 50},
	}

	exporter := xlsxexport.NewReportServerXLSXExporter()
	buf := &bytes.Buffer{}
	err := exporter.Export(context.Background(), buf, data)

	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}

func TestReportServerXLSXExporter_EmptyData(t *testing.T) {
	exporter := xlsxexport.NewReportServerXLSXExporter()
	buf := &bytes.Buffer{}
	err := exporter.Export(context.Background(), buf, []model.ServerUptimeAgg{})

	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}
