package agentruntime

import (
	"fmt"

	agentconfig "github.com/LeHuuHai/server-management/config/agent"
	apperr "github.com/LeHuuHai/server-management/internal/error"
)

type App struct {
	Config *agentconfig.Config
}

func NewApp(cfg *agentconfig.Config) (*App, error) {
	// load config
	cfg, err := agentconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}
	return &App{
		Config: cfg,
	}, nil
}
