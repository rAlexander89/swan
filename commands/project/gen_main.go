package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rAlexander89/swan/utils"
)

func WriteMain(projectPath string) error {
	projectName, err := utils.GetProjectName()
	if err != nil {
		return err
	}

	mainContent := fmt.Sprintf(`package main

import (
    "context"
    "fmt"
    "log"

    "%s/internal/infrastructure/config"
    "%s/internal/infrastructure/server"
)

func main() {
    // load configuration
    env := config.GetEnv()
    cfg, err := config.LoadConfig(env)
    if err != nil {
        log.Fatalf("failed to load config: %%v", err)
    }

    ctx := context.Background()

    // initialize server
    srv, err := server.NewServer(ctx, cfg)
    if err != nil {
        log.Fatalf("failed to initialize server: %%v", err)
    }

    fmt.Printf("starting application in %%s mode...\n", env)

    // start server (blocking)
    if err := srv.Run("8080"); err != nil {
        log.Fatalf("server error: %%v", err)
    }
}`, projectName, projectName)

	mainPath := filepath.Join(projectPath, "cmd", "main.go")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %v", err)
	}

	return nil
}
