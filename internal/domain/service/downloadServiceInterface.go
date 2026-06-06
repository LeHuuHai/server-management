package serviceinterface

import (
	"context"
)

type DownloadServiceInterface interface {
	Download(ctx context.Context, filename string) ([]byte, error)
}
