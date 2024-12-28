package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rAlexander89/swan/utils"
)

const (
	Create = 'C'
	Read   = 'R'
	Update = 'U'
	Delete = 'D'
	Index  = 'I'
)

type routeData struct {
	ProjectName string
	DomainTitle string
	DomainLower string
	DomainSnake string
	Operations  string
}

func WriteRoutes(projectPath, domain, ops string) error {
	// get project name for imports
	projectName, err := utils.GetProjectName()
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	data := routeData{
		ProjectName: projectName,
		DomainTitle: utils.ToUpperFirst(domain),
		DomainLower: strings.ToLower(domain),
		DomainSnake: utils.ToSnakeCase(domain),
		Operations:  ops,
	}

	// ensure routes directory exists
	routesDir := filepath.Join(
		projectPath,
		"internal",
		"infrastructure",
		"server",
		"routes",
		data.DomainSnake,
	)
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		return fmt.Errorf("failed to create routes directory: %v", err)
	}

	// create routes file
	routesPath := filepath.Join(routesDir, fmt.Sprintf("%s_routes.go", data.DomainSnake))

	tmpl := getRoutesTemplate()

	return writeTemplateToFile(routesPath, tmpl, data)
}

func getRoutesTemplate() string {
	return `package {{.DomainLower}}

import (
    "{{.ProjectName}}/internal/app/handlers/api/{{.DomainLower}}s"
    "{{.ProjectName}}/internal/infrastructure/server"
)

type {{.DomainTitle}}Routes struct {
    handler *{{.DomainLower}}s.{{.DomainTitle}}Handler
}

func New{{.DomainTitle}}Routes(handler *{{.DomainLower}}s.{{.DomainTitle}}Handler) *{{.DomainTitle}}Routes {
    return &{{.DomainTitle}}Routes{
        handler: handler,
    }
}

// RegisterRoutes registers the {{.DomainLower}} routes with the given route group
func (r *{{.DomainTitle}}Routes) RegisterRoutes(group *server.RouteGroup) {
    {{if contains .Operations "C"}}
    // POST /{{.DomainLower}}s
    group.POST("/{{.DomainLower}}s", r.handler.Create)
    {{end}}
    {{if contains .Operations "R"}}
    // GET /{{.DomainLower}}s/:id
    group.GET("/{{.DomainLower}}s/:id", r.handler.Get)
    {{end}}
    {{if contains .Operations "U"}}
    // PUT /{{.DomainLower}}s/:id
    group.PUT("/{{.DomainLower}}s/:id", r.handler.Update)
    {{end}}
    {{if contains .Operations "D"}}
    // DELETE /{{.DomainLower}}s/:id
    group.DELETE("/{{.DomainLower}}s/:id", r.handler.Delete)
    {{end}}
    {{if contains .Operations "I"}}
    // GET /{{.DomainLower}}s
    group.GET("/{{.DomainLower}}s", r.handler.Index)
    {{end}}
}`
}

func writeTemplateToFile(path, tmpl string, data routeData) error {
	funcMap := template.FuncMap{
		"contains": func(ops string, op rune) bool {
			return strings.ContainsRune(ops, op)
		},
	}

	t, err := template.New("routes").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return nil
}
