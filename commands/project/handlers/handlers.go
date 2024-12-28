package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rAlexander89/swan/utils"
)

type handlerTemplate struct {
	handler string
}

type templateData struct {
	ProjectName string
	PackageName string
	DomainTitle string
	DomainLower string
	DomainSnake string
	Operations  string
}

type handlerParts struct {
	imports           string
	handlerDefinition string
	constructor       string
	methods           string
	registration      string
}

const (
	Create = 'C'
	Read   = 'R'
	Update = 'U'
	Delete = 'D'
	Index  = 'I'
)

func getHandlerParts(ops string) handlerParts {
	parts := handlerParts{
		imports: `package {{.PackageName}}

import (
    "encoding/json"
    "net/http"
    
    "{{.ProjectName}}/internal/core/domains/{{.DomainLower}}"
    {{.DomainSnake}}_service "{{.ProjectName}}/internal/core/services/{{.DomainSnake}}_service"
    "{{.ProjectName}}/internal/infrastructure/server"
)`,

		handlerDefinition: `
type {{.DomainTitle}}Handler struct {
    service {{.DomainSnake}}_service.Service
}`,

		constructor: `
func New{{.DomainTitle}}Handler(service {{.DomainSnake}}_service.Service) *{{.DomainTitle}}Handler {
    return &{{.DomainTitle}}Handler{
        service: service,
    }
}`,
	}

	// add registration method
	parts.registration = `
// RegisterRoutes implements the RouteRegistrar interface
func (h *{{.DomainTitle}}Handler) RegisterRoutes(group *server.RouteGroup) error {
    basePath := "/{{.DomainLower}}s"`

	// add routes based on operations
	if strings.Contains(ops, string(Create)) {
		parts.methods += `
// Create handles POST requests to create a new {{.DomainLower}}
func (h *{{.DomainTitle}}Handler) Create(w http.ResponseWriter, r *http.Request) {
    var domain{{.DomainTitle}} {{.DomainLower}}.{{.DomainTitle}}
    if err := json.NewDecoder(r.Body).Decode(&domain{{.DomainTitle}}); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    if err := h.service.Create{{.DomainTitle}}(r.Context(), &domain{{.DomainTitle}}); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(domain{{.DomainTitle}})
}`

		parts.registration += `
    group.POST(basePath, h.Create)`
	}

	parts.registration += `
    return nil
}`

	return parts
}

func getHandlerTemplate(ops string) handlerTemplate {
	parts := getHandlerParts(ops)

	return handlerTemplate{
		handler: strings.Join([]string{
			parts.imports,
			parts.handlerDefinition,
			parts.constructor,
			parts.methods,
			parts.registration,
		}, "\n"),
	}
}

func WriteHandler(projectPath, domain, ops string) error {
	projectName, err := utils.GetProjectName()
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	data := templateData{
		ProjectName: projectName,
		PackageName: strings.ToLower(domain),
		DomainTitle: utils.ToUpperFirst(domain),
		DomainLower: strings.ToLower(domain),
		DomainSnake: utils.PascalToSnake(domain),
		Operations:  ops,
	}

	handlerDir := filepath.Join(
		projectPath,
		"internal",
		"infrastructure",
		"http",
		"handlers",
		data.DomainLower+"s",
	)

	if err := os.MkdirAll(handlerDir, 0755); err != nil {
		return fmt.Errorf("failed to create handler directory: %v", err)
	}

	tmpl := getHandlerTemplate(ops)

	if err := writeTemplateToFile(
		filepath.Join(handlerDir, fmt.Sprintf("%s_handler.go", data.DomainLower)),
		tmpl.handler,
		data,
	); err != nil {
		return fmt.Errorf("failed to write handler file: %v", err)
	}

	return nil
}

func writeTemplateToFile(path, tmpl string, data templateData) error {
	t, err := template.New("").Parse(tmpl)
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
