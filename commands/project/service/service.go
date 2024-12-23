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
    "sync"

    "%[1]s/internal/core/domains/%[2]s"
    "%[1]s/internal/core/ports/repository"
)

var (
    instance %[3]sService
    once     sync.Once
)

type %[3]sServiceImpl struct {
    repo repository.%[3]sRepository
}

func New%[3]sService(repo repository.%[3]sRepository) %[3]sService {
    return &%[3]sServiceImpl{
        repo: repo,
    }
}

// Register%[3]sService registers the service with its repository dependency
func Register%[3]sService(repo repository.%[3]sRepository) error {
    if repo == nil {
        return fmt.Errorf("repository cannot be nil")
    }

    once.Do(func() {
        instance = New%[3]sService(repo)
    })

    return nil
}

// Get%[3]sService returns the singleton instance of the service
func Get%[3]sService() %[3]sService {
    return instance
}`, projectName, domainSnake, domain)

	// add CRUD operations
	for _, op := range ops {
		switch op {
		case 'C':
			content += fmt.Sprintf(`

func (s *%[1]sServiceImpl) Create%[1]s(ctx context.Context, %[2]s *%[2]s.%[1]s) error {
    if %[2]s == nil {
        return Err%[1]sInvalid
    }

    if err := s.repo.Create%[1]s(ctx, %[2]s); err != nil {
        return fmt.Errorf("failed to create %[2]s: %%w", err)
    }

    return nil
}`, domain, strings.ToLower(domain))
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

	projectName, pErr := utils.GetProjectName()
	if pErr != nil {
		return pErr
	}

	upperDomain := utils.ToUpperFirst(domain)
	lowerDomain := strings.ToLower(domain)

	var functions []string
	for _, op := range operations {
		fn := fmt.Sprintf(op.function, upperDomain, lowerDomain, lowerDomain, upperDomain)
		functions = append(functions, fn)
	}

	templateStr := `package service

import (
    "context"
    "errors"
    
    "%[1]s/internal/core/domains/%[2]s"
)

var (
    Err%[3]sInvalid = errors.New("invalid %[2]s")
    Err%[3]sExists  = errors.New("%[2]s already exists")
)

type %[3]sService interface {
    %[4]s
}`

	content := fmt.Sprintf(templateStr,
		projectName,                       // [1]
		lowerDomain,                       // [2]
		upperDomain,                       // [3]
		strings.Join(functions, "\n    "), // [4]
	)

	return os.WriteFile(filePath, []byte(content), 0644)
}

func updateExistingInterface(domain, ops, filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse existing file: %v", err)
	}

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

	operations := getOperations(ops)
	for _, op := range operations {
		exists := false
		for _, method := range iface.Methods.List {
			if method.Names[0].Name == fmt.Sprintf("%s%s", op.name, utils.ToUpperFirst(domain)) {
				exists = true
				break
			}
		}

		if !exists {
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

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %v", err)
	}
	defer file.Close()

	return printer.Fprint(file, fset, node)
}
