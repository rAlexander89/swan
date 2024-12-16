package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rAlexander89/swan/utils"
)

func WriteAppModule(projectPath string) error {
	projectName, pErr := utils.GetProjectName()
	if pErr != nil {
		return pErr
	}

	shutdownFuncStr := `
    func (a *App) Shutdown() error {
      if a.postgresDB != nil {
        if err := a.postgresDB.Close(); err != nil {
          return fmt.Errorf("error closing postgres connection: %w", err)
        }
      }
      return nil
    }
  `

	onceFuncStr := `
    once.Do(func() {
      // initialize postgres repository
      postgresConfig := postgres.Config{
      URI:                   cfg.DB.Postgres.URI,
      MaxOpenConnections:    cfg.DB.Postgres.MaxOpenConnections,
      MaxIdleConnections:    cfg.DB.Postgres.MaxIdleConnections,
      MaxConnectionIdleTime: cfg.DB.Postgres.MaxConnectionIdleTime,
      MaxConnectionLifetime: cfg.DB.Postgres.MaxConnectionLifetime,
    }

    postgresDB, err := postgres.NewRepository(ctx, postgresConfig)
    if err != nil {
      initErr = fmt.Errorf("failed to initialize postgres repository: %w", err)
      return
    }

    app = &App{
      config:     cfg,
      postgresDB: postgresDB,
      }
    })
  `
	appContent := fmt.Sprintf(`package app

import (
    "context"
    "sync"
    "fmt"

    "%s/internal/app/repositories/postgres"
    "%s/internal/infrastructure/config"
)

type App struct {
    config     *config.Config
    postgresDB *postgres.Repository
}

var (
    app  *App
    once sync.Once
)

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
    if cfg == nil {
        return nil, fmt.Errorf("config cannot be nil")
    }

    %s

    var initErr error
    
    if initErr != nil {
        return nil, initErr
    }

    return app, nil
}

func (a *App) Config() *config.Config {
    return a.config
}

func (a *App) PostgresDB() *postgres.Repository {
    return a.postgresDB
}

%s

`, projectName, projectName, onceFuncStr, shutdownFuncStr)

	appPath := filepath.Join(projectPath, "internal", "app", "app.go")

	if err := os.WriteFile(appPath, []byte(appContent), 0644); err != nil {
		return fmt.Errorf("failed to write app.go: %v", err)
	}

	return nil
}