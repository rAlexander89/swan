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
	structFields, sErr := getStructFields(domain)
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

	projName, pErr := utils.GetProjectName()
	if pErr != nil {
		return "", pErr
	}

	// interplate values
	// 1. domain string is already PascalCase
	allLower := strings.ToLower(domain)
	packageName := allLower

	snake_case_domain := utils.PascalToSnake(domain)
	fileName := snake_case_domain

	code := fmt.Sprintf(`package %[1]s

import (
    "context"
    "time"
    "%[2]s/internal/core/domains/%[3]s"
)

func (r *Repository) Create%[4]s(ctx context.Context, %[1]s *%[1]s.%[4]s) error {
    query := `+"`"+`
        insert into %[5]s (
            %[6]s
        ) values (
            %[7]s
        )
    `+"`"+`

    now := time.Now().UTC()
    %[1]s.CreatedAt = now
    %[1]s.UpdatedAt = now

    _, err := r.conn.ExecContext(
        ctx,
        query,
        %[8]s,
    )

    return err
}`,
		packageName,                           // [1] package name
		projName,                              // [2] project import path
		fileName,                              // [3] snake_case domain folder
		domain,                                // [4] PascalCase type names
		fmt.Sprintf("%ss", snake_case_domain), // [5] pluralized snake_case table
		columns,                               // [6] column names
		placeholders,                          // [7] sql placeholders
		valueBindings)                         // [8] value bindings

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
