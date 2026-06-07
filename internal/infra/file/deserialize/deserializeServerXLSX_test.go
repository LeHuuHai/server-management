package xlsximport_test

import (
	"bytes"
	"context"
	"testing"

	apperr "github.com/LeHuuHai/server-management/internal/error"
	xlsximport "github.com/LeHuuHai/server-management/internal/infra/file/deserialize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func makeXLSX(t *testing.T, headers []string, rows [][]string) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := "Sheet1"
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}
	buf := &bytes.Buffer{}
	err := f.Write(buf)
	require.NoError(t, err)
	return buf
}

func TestDeserializeServerXLSX_Success(t *testing.T) {
	buf := makeXLSX(t,
		[]string{"Order", "ServerID", "ServerName", "IPv4"},
		[][]string{
			{"1", "s1", "Server1", "1.2.3.4"},
			{"2", "s2", "Server2", "5.6.7.8"},
		},
	)

	importer := xlsximport.NewServerXLSXImporter()
	result, err := importer.Deserialize(context.Background(), buf)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "s1", result[0].ServerID)
	assert.Equal(t, "Server1", result[0].ServerName)
	assert.Equal(t, "1.2.3.4", result[0].IPv4)
	assert.Equal(t, "s2", result[1].ServerID)
}

func TestDeserializeServerXLSX_EmptyRows_ReturnsEmpty(t *testing.T) {
	buf := makeXLSX(t,
		[]string{"Order", "ServerID", "ServerName", "IPv4"},
		[][]string{}, // chỉ có header, không có data
	)

	importer := xlsximport.NewServerXLSXImporter()
	result, err := importer.Deserialize(context.Background(), buf)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDeserializeServerXLSX_RowTooShort_ReturnsInvalidData(t *testing.T) {
	buf := makeXLSX(t,
		[]string{"Order", "ServerID", "ServerName"}, // thiếu cột IPv4
		[][]string{
			{"1", "s1", "Server1"}, // chỉ 3 cột, cần 4
		},
	)

	importer := xlsximport.NewServerXLSXImporter()
	_, err := importer.Deserialize(context.Background(), buf)

	assert.ErrorIs(t, err, apperr.ErrInvalidImportData)
}

func TestDeserializeServerXLSX_InvalidReader_ReturnsError(t *testing.T) {
	importer := xlsximport.NewServerXLSXImporter()
	_, err := importer.Deserialize(context.Background(), bytes.NewBufferString("not-xlsx-content"))

	assert.Error(t, err)
}
