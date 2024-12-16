package db

import (
	"fmt"
	"strings"

	"github.com/rAlexander89/swan/utils"
)

func generateRepositoryInterface(domain string, ops string) (string, error) {
	var methods []string

	varName := strings.ToLower(domain)
	packageName := strings.ToLower(domain)

	for _, op := range ops {
		switch op {
		case Create:
			methods = append(methods, fmt.Sprintf("Create%s(ctx context.Context, %s *%s.%s) error",
				domain, varName, packageName, domain))
		case Read:
			methods = append(methods, fmt.Sprintf("Get%s(ctx context.Context, id string) (*%s.%s, error)",
				domain, packageName, domain))
		case Update:
			methods = append(methods, fmt.Sprintf("Update%s(ctx context.Context, %s *%s.%s) error",
				domain, strings.ToLower(domain), packageName, domain))
		case Delete:
			methods = append(methods, fmt.Sprintf("Delete%s(ctx context.Context, id string) error",
				domain))
		case Index:
			methods = append(methods, fmt.Sprintf("List%ss(ctx context.Context) ([]*%s.%s, error)",
				domain, packageName, domain))
		}
	}

	projectName, pErr := utils.GetProjectName()
	if pErr != nil {
		return "", pErr
	}

	code := fmt.Sprintf(`package %s

import (
    "context"
    "%s/internal/core/domains/%ss"
)

type Repository interface {
    %s
}`, strings.ToLower(domain), projectName, domain, strings.Join(methods, "\n    "))

	return code, nil
}
