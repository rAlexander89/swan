package service

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

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
		// case Read:
		// 	operations = append(operations, operation{
		// 		name:     "Read",
		// 		function: "Get%sByID(ctx context.Context, id string) (*%s.%s, error)",
		// 	})
		// case Index:
		// 	operations = append(operations, operation{
		// 		name:     "Index",
		// 		function: "List%ss(ctx context.Context, filters map[string]interface{}) ([]*%s.%s, error)",
		// 	})
		// }
	}

	return operations
}

func GenerateServiceImpl(domain, ops string) error {
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
	implPath := filepath.Join(serviceDir, fmt.Sprintf("%s_service_impl.go", domainSnake))

	content := fmt.Sprintf(`package service

import (
    "context"
    "fmt"

    "%[1]s/internal/core/domains/%[2]s"
    "%[1]s/internal/core/ports/repository"
)

type %[3]sServiceImpl struct {
    repo repository.%[3]sRepository
}

func New%[3]sService(repo repository.%[3]sRepository) %[3]sService {
    return &%[2]sServiceImpl{
        repo: repo,
    }
}`, projectName, domainSnake, domain)

	for _, op := range ops {
		switch op {
		case 'C':
			content += fmt.Sprintf(`

func (s *%[1]sServiceImpl) Create%[2]s(ctx context.Context, %[1]s *%[1]s.%[2]s) error {
    if %[1]s == nil {
        return Err%[2]sInvalid
    }

    if err := s.repo.Create%[2]s(ctx, %[1]s); err != nil {
        return fmt.Errorf("failed to create %[1]s: %%w", err)
    }

    return nil
}`, domainSnake, domain)
		}
	}

	return os.WriteFile(implPath, []byte(content), 0644)
}

func GenerateServiceInterface(domain, ops string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	serviceDir := filepath.Join(pwd, "internal", "core", "services", fmt.Sprintf("%s_service", utils.PascalToSnake(domain)))
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service directory: %v", err)
	}

	filePath := filepath.Join(serviceDir, fmt.Sprintf("%s_service.go", utils.PascalToSnake(domain)))

	// check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return createNewInterface(domain, ops, filePath)
	}

	return updateExistingInterface(domain, ops, filePath)
}

func createNewInterface(domain, ops, filePath string) error {
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

	projectName, pErr := utils.GetProjectName()
	if pErr != nil {
		return pErr
	}

	content := fmt.Sprintf(`package service

import (
    "context"
    "errors"
    
    "%s/internal/core/domains/%s"
)

var (
    Err%sInvalid = errors.New("invalid %s")
    Err%sExists  = errors.New("%s already exists")
)

type %sService interface {
    %s
}`,
		projectName,
		lowerDomain,
		upperDomain, lowerDomain,
		upperDomain, lowerDomain,
		upperDomain,
		strings.Join(functions, "\n    "))

	return os.WriteFile(filePath, []byte(content), 0644)
}

func updateExistingInterface(domain, ops, filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse existing file: %v", err)
	}

	// find the interface
	var iface *ast.InterfaceType
	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == fmt.Sprintf("%sService", utils.ToUpperFirst(domain)) {
				if i, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					iface = i
					return false
				}
			}
		}
		return true
	})

	if iface == nil {
		return fmt.Errorf("interface not found in file")
	}

	// add new methods
	operations := getOperations(ops)
	for _, op := range operations {
		// check if method already exists
		exists := false
		for _, method := range iface.Methods.List {
			if method.Names[0].Name == fmt.Sprintf("%s%s", op.name, utils.ToUpperFirst(domain)) {
				exists = true
				break
			}
		}

		if !exists {
			// parse new method
			expr := fmt.Sprintf("type T interface { %s }",
				fmt.Sprintf(op.function,
					utils.ToUpperFirst(domain),
					strings.ToLower(domain),
					domain,
					utils.ToUpperFirst(domain)))

			f, err := parser.ParseFile(fset, "", expr, 0)
			if err != nil {
				return fmt.Errorf("failed to parse new method: %v", err)
			}

			typeSpec := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
			ifaceType := typeSpec.Type.(*ast.InterfaceType)
			iface.Methods.List = append(iface.Methods.List, ifaceType.Methods.List...)
		}
	}

	// write updated file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %v", err)
	}
	defer file.Close()

	return printer.Fprint(file, fset, node)
}
