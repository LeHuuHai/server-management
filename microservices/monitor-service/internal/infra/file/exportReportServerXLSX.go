package file

import (
	"context"
	"io"
	"strconv"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"github.com/xuri/excelize/v2"
)

type ReportExporter struct{}

func NewReportExporter() *ReportExporter {
	return &ReportExporter{}
}

func (e *ReportExporter) FileType() string {
	return "xlsx"
}

func (e *ReportExporter) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

func (e *ReportExporter) Export(ctx context.Context, writer io.Writer, data []model.ServerUptimeAgg) error {
	f := excelize.NewFile()
	sheet := "Servers"
	f.SetSheetName("Sheet1", sheet)
	f.SetCellValue(sheet, "A1", "Order number")
	f.SetCellValue(sheet, "B1", "ServerID")
	f.SetCellValue(sheet, "C1", "Uptime Ratio")
	f.SetCellValue(sheet, "D1", "Start Ping At")
	f.SetCellValue(sheet, "E1", "Last Ping At")
	f.SetCellValue(sheet, "F1", "Total Checks")

	for idx, item := range data {
		row := strconv.Itoa(idx + 2)
		f.SetCellValue(sheet, "A"+row, idx+1)
		f.SetCellValue(sheet, "B"+row, item.ServerID)
		f.SetCellValue(sheet, "C"+row, item.UptimeRatio)
		f.SetCellValue(sheet, "D"+row, item.StartPingAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "E"+row, item.LastPingAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "F"+row, item.DocCount)
	}

	return f.Write(writer)
}
