package db

import (
	"fmt"
	"strings"

	"github.com/rAlexander89/swan/utils"
)

func generateRepositoryInterface(domain string, ops string) (string, error) {
	if domain == "" {
		return "", fmt.Errorf("domain name cannot be empty")
	}

	// get project name for imports
	projectName, err := utils.GetProjectName()
	if err != nil {
		return "", fmt.Errorf("failed to get project name: %v", err)
	}

	// normalize domain names for different uses
	domainLower := strings.ToLower(domain)
	domainTitle := utils.ToUpperFirst(domain)

	// generate repository interface
	code := fmt.Sprintf(`package %s

import (
    "context"
    "%s/internal/core/domains/%s"
)

// repository interface for %s domain
type Repository interface {
    %s
}`, domainLower, projectName, domainLower, domainTitle,
		strings.Join(buildMethodList(domainTitle, domainLower, ops), "\n    "))

	return code, nil
}

func buildMethodList(domainTitle, domainLower, ops string) []string {
	var methods []string

	for _, op := range ops {
		switch op {
		case Create:
			methods = append(methods, fmt.Sprintf(
				"Create%s(ctx context.Context, %s *%s.%s) error",
				domainTitle, domainLower, domainLower, domainTitle,
			))
		case Read:
			methods = append(methods, fmt.Sprintf(
				"Get%s(ctx context.Context, id string) (*%s.%s, error)",
				domainTitle, domainLower, domainTitle,
			))
		case Update:
			methods = append(methods, fmt.Sprintf(
				"Update%s(ctx context.Context, %s *%s.%s) error",
				domainTitle, domainLower, domainLower, domainTitle,
			))
		case Delete:
			methods = append(methods, fmt.Sprintf(
				"Delete%s(ctx context.Context, id string) error",
				domainTitle,
			))
		case Index:
			methods = append(methods, fmt.Sprintf(
				"List%ss(ctx context.Context) ([]*%s.%s, error)",
				domainTitle, domainLower, domainTitle,
			))
		}
	}

	return methods
}
