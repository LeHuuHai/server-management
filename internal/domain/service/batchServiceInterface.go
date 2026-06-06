package serviceinterface

import "context"

type BatchServiceInterface interface {
	Run(ctx context.Context)
}
