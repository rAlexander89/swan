// commands/fly/fly.go
package fly

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	project "github.com/rAlexander89/swan/commands/project/handlers"
	routes "github.com/rAlexander89/swan/commands/project/routes"
	"github.com/rAlexander89/swan/nodes"
	"github.com/rAlexander89/swan/utils"
)

func init() {
	nodes.RegisterCommand("fly", Fly)
}

func Fly(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("domain name required")
	}

	domain := args[0]
	var ops string

	// check for operations flag
	for i, arg := range args {
		if arg == "-c" && i+1 < len(args) {
			ops = strings.ToUpper(args[i+1])

			// for now, only support Create
			if !strings.Contains(ops, "C") {
				return fmt.Errorf("currently only Create operation is supported")
			}
			break
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	// check if domain struct exists
	domainPath := filepath.Join(
		pwd,
		"internal",
		"core",
		"domains",
		strings.ToLower(domain),
		fmt.Sprintf("%s.go", utils.PascalToSnake(domain)),
	)

	if _, err := os.Stat(domainPath); os.IsNotExist(err) {
		return fmt.Errorf("domain %s not found at %s", domain, domainPath)
	}

	// generate handler
	if err := project.WriteHandler(".", domain, ops); err != nil {
		return fmt.Errorf("error generating handler: %v", err)
	}

	// generate routes
	if err := routes.WriteRoutes(".", domain, ops); err != nil {
		return fmt.Errorf("error generating routes: %v", err)
	}

	// register with api routes
	if err := routes.WriteTopLevelRoutes(".", domain, ops); err != nil {
		return fmt.Errorf("error registering routes: %v", err)
	}

	return nil
}
