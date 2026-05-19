package export

type FileType string

const (
	FileXLSX FileType = "xlsx"
)

type DataType string

const (
	DataServer DataType = "server"
)

type ExportFunc func([]any) any

type Exporter struct {
	FileType   FileType
	DataType   DataType
	ExportFunc ExportFunc
}
