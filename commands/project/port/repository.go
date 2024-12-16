package port

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rAlexander89/swan/utils"
)

func GenerateRepositoryPort(domain string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	repoDir := filepath.Join(pwd, "internal", "core", "ports", "repository")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %v", err)
	}

	projName, pErr := utils.GetProjectName()
	if pErr != nil {
		return pErr
	}

	data := struct {
		Proj        string
		Domain      string
		LowerDomain string
	}{
		Proj:        projName,
		Domain:      utils.ToUpperFirst(domain),
		LowerDomain: strings.ToLower(domain),
	}

	tmpl := template.Must(template.New("repository").Parse(`package repository

import (
    "context"
    "errors"
    
    "{{.Proj}}/internal/core/domains/{{.LowerDomain}}"
)

var (
    Err{{.Domain}}NotCreated = errors.New("failed to create {{.LowerDomain}}")
)

type {{.Domain}}Repository interface {
    Create{{.Domain}}(ctx context.Context, {{.LowerDomain}} *{{.LowerDomain}}.{{.Domain}}) error
}`))

	filePath := filepath.Join(repoDir, fmt.Sprintf("%s_repository.go", utils.PascalToSnake(domain)))

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create repository file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write repository template: %v", err)
	}

	return nil
}
