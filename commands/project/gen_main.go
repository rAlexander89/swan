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
    "os"
    "os/signal"
    "syscall"
    "time"

    "%s/internal/app"
    "%s/internal/infrastructure/config"
)

func main() {
    // load configuration
    env := config.GetEnv()
    cfg, err := config.LoadConfig(env)
    if err != nil {
        log.Fatalf("failed to load config: %%v", err)
    }

    // setup context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // initialize application
    app, err := app.NewApp(ctx, cfg)
    if err != nil {
        log.Fatalf("failed to initialize application: %%v", err)
    }
    defer app.Shutdown()

    // handle graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        sig := <-quit
        log.Printf("received signal: %%v", sig)
        cancel()

        // force shutdown after timeout
        time.Sleep(10 * time.Second)
        log.Fatal("forced shutdown after timeout")
    }()

    fmt.Printf("starting application in %%s mode...\n", env)

    // start application (blocking)
    if err := app.Start(ctx); err != nil {
        log.Fatalf("application error: %%v", err)
    }
}`, projectName, projectName)

	mainPath := filepath.Join(projectPath, "cmd", "main.go")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %v", err)
	}

	return nil
}
