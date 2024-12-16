package db

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/rAlexander89/swan/utils"
)

func generateCreate(domain string) (string, error) {
	structFields, sErr := getStructFields(domain) // We'll implement this next
	if sErr != nil {
		return "", fmt.Errorf("error reading struct fields for %s: %v ", domain, sErr)
	}

	columns := make([]string, 0, len(structFields))
	placeholders := make([]string, 0, len(structFields))
	valueBindings := make([]string, 0, len(structFields))

	for i, field := range structFields {
		columns = append(columns, utils.ToSnakeCase(field.Name))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		valueBindings = append(valueBindings, fmt.Sprintf("%s.%s", strings.ToLower(domain), field.Name))
	}

	code := fmt.Sprintf(`package %s

import (
   "context"
   "time"
   "your_project/internal/core/domains/%s"
)

func (r *Repository) Create%s(ctx context.Context, %s *%s.%s) error {
   query := `+"`"+`
       insert into %ss (
           %s
       ) values (
           %s
       )
   `+"`"+`

   now := time.Now().UTC()
   %s.CreatedAt = now
   %s.UpdatedAt = now

   _, err := r.conn.ExecContext(
       ctx,
       query,
       %s,
   )

   return err
}`,
		domain, domain,
		domain,
		strings.ToLower(domain), domain, domain,
		strings.ToLower(domain),
		strings.Join(columns, ",\n            "),
		strings.Join(placeholders, ",\n            "),
		strings.ToLower(domain),
		strings.ToLower(domain),
		strings.Join(valueBindings, ",\n        "))

	return code, nil
}

type Field struct {
	Name string
	Type string
	Tags map[string]string
}

func getStructFields(domain string) ([]Field, error) { // ex User
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}

	// convert domain name to snake_case for file path
	domainPath := filepath.Join(
		pwd,
		"internal",
		"core",
		"domains",
		strings.ToLower(domain),
		fmt.Sprintf("%s.go", utils.PascalToSnake(domain)),
	)

	content, err := os.ReadFile(domainPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read domain file: %v", err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", string(content), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse domain file: %v", err)
	}

	var fields []Field
	ast.Inspect(f, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// check if this is our domain struct
		if typeSpec.Name.Name != domain {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// extract fields from struct
		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				continue
			}

			fields = append(fields, Field{
				Name: field.Names[0].Name,
				Type: getFieldType(field.Type),
				Tags: parseStructTags(field.Tag),
			})
		}

		return false
	})

	return fields, nil
}

func parseStructTags(tag *ast.BasicLit) map[string]string {
	if tag == nil {
		return nil
	}

	tags := make(map[string]string)
	tagStr := strings.Trim(tag.Value, "`")

	for _, tag := range strings.Split(tagStr, " ") {
		parts := strings.Split(tag, ":")
		if len(parts) != 2 {
			continue
		}
		key := strings.Trim(parts[0], "\"")
		value := strings.Trim(parts[1], "\"")
		tags[key] = value
	}

	return tags
}

func getFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		x, ok := t.X.(*ast.Ident)
		if !ok {
			return ""
		}
		return fmt.Sprintf("%s.%s", x.Name, t.Sel.Name)
	case *ast.StarExpr:
		return "*" + getFieldType(t.X)
	case *ast.ArrayType:
		return "[]" + getFieldType(t.Elt)
	default:
		return ""
	}
}
