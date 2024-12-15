package domain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rAlexander89/swan/utils"
)

// creates a new domain

// possible args

type domainArgs struct {
	name     string
	argType  string
	required bool
}

func Create(args []string) error {
	argL := len(args)
	if argL == 0 {
		return errors.New("expected at least 1 argument: domain name")
	}

	// new domain
	var fields []utils.Field
	var tags []string

	// start @ 1 to validate command args
	for i := 1; i < len(args); i++ {
		arg := args[i]
		var err error

		switch {
		case strings.HasPrefix(arg, "-f"):
			fields, err = utils.ParseFields(strings.TrimPrefix(arg, "-f="))
			if err != nil {
				return fmt.Errorf("failed to parse fields: %v ", err)
			}
		case strings.HasPrefix(arg, "-t"):
			tags, err = utils.ParseTags(strings.TrimPrefix(arg, "-t="))
			if err != nil {
				return fmt.Errorf("failed to parse tags: %v", err)
			}
		}
	}

	domain := strings.ToLower(args[0])
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	domainPath := filepath.Join(currentDir, "internal", "core", "domains", domain)

	if err := os.MkdirAll(domainPath, 0755); err != nil {
		return fmt.Errorf("failed to create domain directory: %v", err)
	}

	structFields := ""
	for _, f := range fields {
		tagStr := utils.GenerateTags(f.Name, tags)
		structFields += fmt.Sprintf("    %s %s `%s`\n", f.Name, f.DataType, tagStr)
	}

	// gen struct
	// create domain file content
	domainContent := fmt.Sprintf(
		`// %s.go
  package %sDomain

  type %s struct {
    %s
  }
  `,
		utils.ToSnakeCase(domain),   // domain_name.go
		utils.PascalToLower(domain), // pacakge domainname
		domain,                      // type PublicDomain struct
		structFields,                // struct fields
	)

	// write domain file
	domainFile := filepath.Join(domainPath, domain+".go")
	if err := os.WriteFile(domainFile, []byte(domainContent), 0644); err != nil {
		return fmt.Errorf("failed to create domain file: %v", err)
	}

	return nil
}
