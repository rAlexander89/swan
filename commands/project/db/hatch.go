// commands/db/hatch.go
package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rAlexander89/swan/commands/project/port"
	"github.com/rAlexander89/swan/commands/project/service"
	"github.com/rAlexander89/swan/nodes"
	"github.com/rAlexander89/swan/utils"
)

// operation flags
const (
	Create = 'C'
	Read   = 'R'
	Update = 'U'
	Delete = 'D'
	Index  = 'I'
)

func init() {
	nodes.RegisterCommand("hatch", Hatch)
}

type operation struct {
	name     string
	filename string
	content  string
}

func Hatch(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("domain name required")
	}

	domain := args[0]
	ops := "CRUDI"

	// check for operations flag
	for i, arg := range args {
		if arg == "-c" && i+1 < len(args) {
			ops = strings.ToUpper(args[i+1])
			break
		}
	}

	// validate operations
	for _, op := range ops {
		switch op {
		case Create, Read, Update, Delete, Index:
			continue
		default:
			return fmt.Errorf("invalid operation: %c", op)
		}
	}

	// get project root and create paths
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	domain_snake := utils.PascalToSnake(domain)

	// 1. postgres repository implementation
	repoPath := filepath.Join(pwd, "internal", "app", "repositories", "postgres", "domains", domain_snake)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %v", err)
	}

	// generate files based on operations
	operations := []operation{}

	persistenceContnent, pErr := generateRepositoryInterface(domain, ops)
	if pErr != nil {
		return pErr
	}

	// always create repository interface
	operations = append(operations, operation{
		name:     "repository",
		filename: fmt.Sprintf("%s_repository.go", domain_snake),
		content:  persistenceContnent,
	})

	// add operation files
	for _, op := range ops {
		switch op {
		case Create:

			content, cErr := generateCreate(domain)
			if cErr != nil {
				return scaffoldErr(domain, op)
			}

			operations = append(operations, operation{
				name:     "create",
				filename: fmt.Sprintf("%s_create.go", domain_snake),
				content:  content,
			})
			// case Read:
			// 	operations = append(operations, operation{
			// 		name:     "read",
			// 		filename: fmt.Sprintf("%s_read.go", domain),
			//    content:  generateGet(domain),
			// 	})
			// case Update:
			// 	operations = append(operations, operation{
			// 		name:     "update",
			// 		filename: fmt.Sprintf("%s_update.go", domain),
			// 		content:  generateUpdate(domain),
			// 	})
			// case Delete:
			// 	operations = append(operations, operation{
			// 		name:     "delete",
			// 		filename: fmt.Sprintf("%s_delete.go", domain),
			// 		content:  generateDelete(domain),
			// 	})
			// case Index:
			// 	operations = append(operations, operation{
			// 		name:     "index",
			// 		filename: fmt.Sprintf("%s_index.go", domain),
			// 		content:  generateIndex(domain),
			// 	})
		}
	}

	// writes postgres > domain_repository file
	for _, op := range operations {
		path := filepath.Join(repoPath, op.filename)
		if err := os.WriteFile(path, []byte(op.content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", op.name, err)
		}
	}

	// generate port > repository > domain file
	if err := port.GenerateRepositoryPort(domain); err != nil {
		return err
	}

	// 2. repository port interface
	if err := port.GenerateRepositoryPort(domain); err != nil {
		return fmt.Errorf("failed to generate repository port: %v", err)
	}

	// 3. generate domains services
	if err := service.GenerateService(domain, ops); err != nil {
		return fmt.Errorf("failed to generate service: %v", err)
	}

	return nil
}

func scaffoldErr(domain string, op rune) error {
	errStr := "error scaffoling %s file for %s operation"
	return fmt.Errorf(errStr, domain, op)
}
