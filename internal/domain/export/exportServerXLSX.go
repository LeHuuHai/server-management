package export

import (
	"strconv"

	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/xuri/excelize/v2"
)

func ExportServerXLSX(data []model.Server) (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "Servers"
	f.SetSheetName("Sheet1", sheet)
	f.SetCellValue(sheet, "A1", "Order number")
	f.SetCellValue(sheet, "B1", "ServerID")
	f.SetCellValue(sheet, "C1", "ServerName")
	f.SetCellValue(sheet, "D1", "Ipv4")
	f.SetCellValue(sheet, "E1", "Status")
	f.SetCellValue(sheet, "F1", "CreateAt")
	f.SetCellValue(sheet, "G1", "MetadataUpdatedAt")
	f.SetCellValue(sheet, "H1", "LastPingAt")

	for idx, item := range data {
		row := strconv.Itoa(idx + 2)
		f.SetCellValue(sheet, "A"+row, idx+1)
		f.SetCellValue(sheet, "B"+row, item.ServerID)
		f.SetCellValue(sheet, "C"+row, item.ServerName)
		f.SetCellValue(sheet, "D"+row, item.IPv4)
		f.SetCellValue(sheet, "E"+row, item.Status)
		f.SetCellValue(sheet, "F"+row, item.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "G"+row, item.MetadataUpdatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "H"+row, item.LastPingAt.Format("2006-01-02 15:04:05"))
	}

	return f, nil
}
