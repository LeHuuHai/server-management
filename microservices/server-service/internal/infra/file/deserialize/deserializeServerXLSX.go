package xlsximport

import (
	"context"
	"io"

	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"github.com/xuri/excelize/v2"
)

type serverXLSXImporter struct{}

func (i *serverXLSXImporter) Deserialize(ctx context.Context, reader io.Reader) ([]model.ServerImport, error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, apperr.ErrInvalidImportData
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	if len(rows) <= 1 {
		return []model.ServerImport{}, nil
	}

	servers := make([]model.ServerImport, 0, len(rows)-1)

	for _, row := range rows[1:] {

		if len(row) < 4 {
			return nil, apperr.ErrInvalidImportData
		}

		server := model.ServerImport{
			ServerID:   row[1],
			ServerName: row[2],
			IPv4:       row[3],
		}

		servers = append(servers, server)
	}

	return servers, nil
}

func NewServerXLSXImporter() *serverXLSXImporter {
	return &serverXLSXImporter{}
}
