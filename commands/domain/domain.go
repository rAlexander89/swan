package domain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rAlexander89/swan/nodes"
	"github.com/rAlexander89/swan/utils"
)

func init() {
	nodes.RegisterCommand("domain", Create)
}

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

	domain := args[0] // SomeDomain
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	fileName := utils.PascalToSnake(domain) // Some_Domain
	fileName = strings.ToLower(fileName)    // some_domain
	domainPath := filepath.Join(currentDir, "internal", "core", "domains", domain)

	fmt.Printf("generating new domain %s", domain)

	// validate struct fields
	var fields []utils.Field
	var tags []string

	// start @ 1. index 0 is the domain name
	for i := 1; i < len(args); i++ {
		arg := args[i]
		var err error

		switch arg {
		case "-f":
			fmt.Println("generating struct fields")
			fields, err = utils.ParseArgFields(args, i+1)
			if err != nil {
				return fmt.Errorf("failed to parse fields: %v ", err)
			}
		case "-t":
			fmt.Println("generating struct field tags")
			tags, err = utils.ParseArgTags(args, i+1)
			if err != nil {
				return fmt.Errorf("failed to parse tags: %v", err)
			}
		}
	}

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
  package %s

  type %s struct {
    %s
  }
  `,
		fileName,                // domain_name.go
		strings.ToLower(domain), // pacakge domainname
		domain,                  // type PublicDomain struct
		structFields,            // struct fields
	)

	// write domain file
	domainFile := filepath.Join(domainPath, domain+".go")
	if err := os.WriteFile(domainFile, []byte(domainContent), 0644); err != nil {
		return fmt.Errorf("failed to create domain file: %v", err)
	}

	return nil
}
