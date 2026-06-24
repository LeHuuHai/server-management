package xlsxexport

import (
	"context"
	"io"
	"strconv"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"github.com/xuri/excelize/v2"
)

type serverXLSXExporter struct{}

func (e *serverXLSXExporter) FileType() string {
	return "xlsx"
}

func (e *serverXLSXExporter) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

func (e *serverXLSXExporter) Export(ctx context.Context, writer io.Writer, data []model.ServerProfile) error {
	f := excelize.NewFile()
	sheet := "Servers"
	f.SetSheetName("Sheet1", sheet)
	f.SetCellValue(sheet, "A1", "Order number")
	f.SetCellValue(sheet, "B1", "ServerID")
	f.SetCellValue(sheet, "C1", "ServerName")
	f.SetCellValue(sheet, "D1", "Ipv4")
	f.SetCellValue(sheet, "E1", "Status")
	f.SetCellValue(sheet, "F1", "CreateAt")
	f.SetCellValue(sheet, "G1", "UpdatedAt")
	f.SetCellValue(sheet, "H1", "LastPingAt")

	for idx, item := range data {
		row := strconv.Itoa(idx + 2)
		f.SetCellValue(sheet, "A"+row, idx+1)
		f.SetCellValue(sheet, "B"+row, item.ServerID)
		f.SetCellValue(sheet, "C"+row, item.ServerName)
		f.SetCellValue(sheet, "D"+row, item.IPv4)
		f.SetCellValue(sheet, "E"+row, "UNKNOWN")
		f.SetCellValue(sheet, "F"+row, item.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "G"+row, item.UpdatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "H"+row, "")
	}

	return f.Write(writer)
}

func NewServerXLSXExporter() *serverXLSXExporter {
	return &serverXLSXExporter{}
}
