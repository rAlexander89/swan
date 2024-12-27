// commands/project/service/service.go
package service

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

type operation struct {
	name     string
	function string
}

func getOperations(ops string) []operation {
	operations := []operation{}
	for _, op := range ops {
		switch op {
		case Create:
			operations = append(operations, operation{
				name:     "Create",
				function: "Create%s(ctx context.Context, %s *%s.%s) error",
			})
		}
	}
	return operations
}

func GenerateService(domain, ops string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	projectName, err := utils.GetProjectName()
	if err != nil {
		return err
	}

	domainSnake := utils.PascalToSnake(domain)
	serviceDir := filepath.Join(pwd, "internal", "core", "services", fmt.Sprintf("%s_service", domainSnake))

	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service directory: %v", err)
	}

	// generate types.go with interface and errors
	if err := generateTypes(domain, ops, serviceDir, projectName); err != nil {
		return fmt.Errorf("failed to generate types: %v", err)
	}

	// generate domain.go with implementation
	if err := generateImplementation(domain, ops, serviceDir, projectName); err != nil {
		return fmt.Errorf("failed to generate implementation: %v", err)
	}

	return nil
}

func generateTypes(domain, ops, serviceDir, projectName string) error {
	operations := getOperations(ops)
	if len(operations) == 0 {
		return fmt.Errorf("no valid operations provided")
	}

	upperDomain := utils.ToUpperFirst(domain)
	lowerDomain := strings.ToLower(domain)

	var functions []string
	for _, op := range operations {
		fn := fmt.Sprintf(op.function, upperDomain, lowerDomain, domain, upperDomain)
		functions = append(functions, fn)
	}

	tmpl := template.Must(template.New("types").Parse(`package {{.Package}}

import (
    "context"
    "errors"
    
    "{{.ProjectName}}/internal/core/domains/{{.DomainLower}}"
)

var (
    Err{{.DomainUpper}}Invalid = errors.New("invalid {{.DomainLower}}")
    Err{{.DomainUpper}}Exists  = errors.New("{{.DomainLower}} already exists")
)

type Service interface {
    {{range .Functions}}
    {{.}}{{end}}
}`))

	data := struct {
		Package     string
		ProjectName string
		DomainUpper string
		DomainLower string
		Functions   []string
	}{
		Package:     fmt.Sprintf("%s_service", lowerDomain),
		ProjectName: projectName,
		DomainUpper: upperDomain,
		DomainLower: lowerDomain,
		Functions:   functions,
	}

	typesFile, err := os.Create(filepath.Join(serviceDir, "types.go"))
	if err != nil {
		return err
	}
	defer typesFile.Close()

	return tmpl.Execute(typesFile, data)
}

func generateImplementation(domain, ops, serviceDir, projectName string) error {
	upperDomain := utils.ToUpperFirst(domain)
	lowerDomain := strings.ToLower(domain)

	tmpl := template.Must(template.New("impl").Parse(`package {{.Package}}

import (
    "context"
    "fmt"

    "{{.ProjectName}}/internal/core/domains/{{.DomainLower}}"
    "{{.ProjectName}}/internal/core/ports/repository"
)

type service struct {
    repo repository.{{.DomainUpper}}Repository
}

func New(repo repository.{{.DomainUpper}}Repository) Service {
    return &service{
        repo: repo,
    }
}

{{range .Operations}}
func (s *service) Create{{$.DomainUpper}}(ctx context.Context, {{$.DomainLower}} *{{$.DomainLower}}.{{$.DomainUpper}}) error {
    if {{$.DomainLower}} == nil {
        return Err{{$.DomainUpper}}Invalid
    }

    if err := s.repo.Create{{$.DomainUpper}}(ctx, {{$.DomainLower}}); err != nil {
        return fmt.Errorf("failed to create {{$.DomainLower}}: %w", err)
    }

    return nil
}{{end}}`))

	data := struct {
		Package     string
		ProjectName string
		DomainUpper string
		DomainLower string
		Operations  []operation
	}{
		Package:     fmt.Sprintf("%s_service", lowerDomain),
		ProjectName: projectName,
		DomainUpper: upperDomain,
		DomainLower: lowerDomain,
		Operations:  getOperations(ops),
	}

	implFile, err := os.Create(filepath.Join(serviceDir, fmt.Sprintf("%s.go", lowerDomain)))
	if err != nil {
		return err
	}
	defer implFile.Close()

	return tmpl.Execute(implFile, data)
}
