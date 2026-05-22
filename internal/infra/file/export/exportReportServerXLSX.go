package xlsxexport

import (
	"context"
	"io"
	"strconv"

	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/xuri/excelize/v2"
)

type reportServerXLSXExport struct{}

func (e *reportServerXLSXExport) FileType() string {
	return "xlsx"
}

func (e *reportServerXLSXExport) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

func (e *reportServerXLSXExport) Export(ctx context.Context, writer io.Writer, data []model.ServerUptimeAgg) error {
	f := excelize.NewFile()
	sheet := "Servers"
	f.SetSheetName("Sheet1", sheet)
	f.SetCellValue(sheet, "A1", "Order number")
	f.SetCellValue(sheet, "B1", "ServerID")
	f.SetCellValue(sheet, "C1", "Uptime Ratio")
	f.SetCellValue(sheet, "D1", "Start Ping At")
	f.SetCellValue(sheet, "E1", "Last Ping At")

	for idx, item := range data {
		row := strconv.Itoa(idx + 2)
		f.SetCellValue(sheet, "A"+row, idx+1)
		f.SetCellValue(sheet, "B"+row, item.ServerID)
		f.SetCellValue(sheet, "C"+row, item.UptimeRatio)
		f.SetCellValue(sheet, "D"+row, item.StartPingAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "E"+row, item.LastPingAt.Format("2006-01-02 15:04:05"))
	}

	return f.Write(writer)
}

func NewReportServerXLSXExporter() *reportServerXLSXExport {
	return &reportServerXLSXExport{}
}
