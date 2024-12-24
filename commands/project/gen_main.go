package project

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/rAlexander89/swan/utils"
)

func WriteMain(projectPath string) error {
	projectName, err := utils.GetProjectName()
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	mainTmpl := `package main

import (
   "context"
   "fmt"
   "log"
   "os"
   "os/signal"
   "syscall"
   "time"

   "{{.ProjectName}}/internal/app"
   "{{.ProjectName}}/internal/infrastructure/config"
   "{{.ProjectName}}/internal/infrastructure/server"
   "{{.ProjectName}}/internal/app/handlers/api"
)

func main() {
   // load configuration
   env := config.GetEnv()
   cfg, err := config.LoadConfig(env)
   if err != nil {
       log.Fatalf("failed to load config: %v", err)
   }

   ctx := context.Background()

   // initialize app with services
   app, err := app.NewApp(ctx, cfg)
   if err != nil {
       log.Fatalf("failed to initialize app: %v", err)
   }
   defer app.Shutdown()

   // initialize server
   srv, err := server.NewServer(ctx, cfg)
   if err != nil {
       log.Fatalf("failed to initialize server: %v", err)
   }

   // register routes with server
   apiGroup := srv.Group("/api")
   v1Group := apiGroup.Group("/v1")

   fmt.Printf("starting application in %s mode...\n", env)

   // start server (blocking)
   if err := srv.Run("8080"); err != nil {
       log.Fatalf("server error: %v", err)
   }
}`

	data := struct {
		ProjectName string
	}{
		ProjectName: projectName,
	}

	mainPath := filepath.Join(projectPath, "cmd", "main.go")

	// ensure directory exists
	if err := os.MkdirAll(filepath.Dir(mainPath), 0755); err != nil {
		return fmt.Errorf("failed to create main directory: %v", err)
	}

	tmpl, err := template.New("main").Parse(mainTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse main template: %v", err)
	}

	f, err := os.Create(mainPath)
	if err != nil {
		return fmt.Errorf("failed to create main.go: %v", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to write main template: %v", err)
	}

	return nil
}
