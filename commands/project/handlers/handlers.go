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
    "{{.ProjectName}}/internal/core/services/{{.DomainSnake}}_service/service"
)`,

		handlerDefinition: `
type {{.DomainTitle}}Handler struct {
    service service.{{.DomainTitle}}Service
}`,

		constructor: `
func New{{.DomainTitle}}Handler(service service.{{.DomainTitle}}Service) *{{.DomainTitle}}Handler {
    return &{{.DomainTitle}}Handler{
        service: service,
    }
}`,
	}

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
	}

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
		"app",
		"handlers",
		"api",
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