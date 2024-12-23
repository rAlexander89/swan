package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rAlexander89/swan/utils"
)

func WriteTopLevelRoutes(projectPath, domain, ops string) error {
	// get project name for imports
	projectName, err := utils.GetProjectName()
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	data := struct {
		ProjectName string
		DomainTitle string
		DomainLower string
	}{
		ProjectName: projectName,
		DomainTitle: utils.ToUpperFirst(domain),
		DomainLower: strings.ToLower(domain),
	}

	// ensure routes directory exists
	routesDir := filepath.Join(projectPath, "internal", "infrastructure", "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		return fmt.Errorf("failed to create routes directory: %v", err)
	}

	// create routes.go file
	routesPath := filepath.Join(routesDir, "routes.go")
	if _, err := os.Stat(routesPath); os.IsNotExist(err) {
		// routes.go doesn't exist, create it
		tmpl := template.Must(template.New("routes").Parse(`package routes

import (
    "{{.ProjectName}}/internal/app/handlers/api/{{.DomainLower}}s"
    "{{.ProjectName}}/internal/infrastructure/server"
    "{{.ProjectName}}/internal/infrastructure/routes/{{.DomainLower}}"
)

func RegisterRoutes(s *server.Server, {{.DomainLower}}Handler *{{.DomainLower}}s.{{.DomainTitle}}Handler) {
    apiGroup := s.Group("/api")
    v1Group := apiGroup.Group("/v1")

    // register {{.DomainLower}} routes
    {{.DomainLower}}Routes := {{.DomainLower}}.New{{.DomainTitle}}Routes({{.DomainLower}}Handler)
    {{.DomainLower}}Routes.RegisterRoutes(v1Group)

    // register other domain routes as needed
    // 
}`))

		f, err := os.Create(routesPath)
		if err != nil {
			return fmt.Errorf("failed to create routes.go file: %v", err)
		}
		defer f.Close()

		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("failed to execute routes.go template: %v", err)
		}
	}

	return nil
}
