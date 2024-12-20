// commands/project/new.go
package project

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rAlexander89/swan/commands/project/server"
	"github.com/rAlexander89/swan/nodes"
)

func init() {
	nodes.RegisterCommand("new", New)
}

//go:embed new.json
var projectStructure []byte

type FileStructure struct {
	Type     string                   `json:"type"`
	Children map[string]FileStructure `json:"children,omitempty"`
}

// creates a new project directory and initializes a go module
func New(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments: directory name and project name")
	}

	dirName := args[0]
	projectName := args[1]

	// get gopath
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return fmt.Errorf("GOPATH environment variable not set")
	}

	// create full project path
	projectPath := filepath.Join(gopath, dirName)

	// check if directory already exists
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return fmt.Errorf("directory already exists: %s", projectPath)
	}

	// create project directory
	err := os.MkdirAll(projectPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// change to project directory
	err = os.Chdir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to change directory: %v", err)
	}

	// initialize go module
	cmd := exec.Command("go", "mod", "init", projectName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize go module: %v\noutput: %s", err, string(output))
	}

	// scaffold project directories
	err = ScaffoldDirs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to scaffold project directories: %v", err)
	}

	// write config env.jsons and config structs
	if err := WriteConfig(projectPath); err != nil {
		return fmt.Errorf("failed to write config files: %v", err)
	}

	if err := WriteConfigLoader(projectPath); err != nil {
		return fmt.Errorf("failed to write config loader file: %v", err)
	}

	if err := WriteMain(projectPath); err != nil {
		return fmt.Errorf("failed to write main.go: %v", err)
	}

	if err := WriteAppModule(projectPath); err != nil {
		return fmt.Errorf("failed to write app.go: %v", err)
	}

	if err := server.WriteServer(projectPath); err != nil {
		return fmt.Errorf("failed to write server.go: %v", err)
	}

	// before go mod init

	fmt.Printf("successfully created new project at %s\n", projectPath)
	return nil
}

// ScaffoldDirs creates the directory structure and files based on embedded new.json
func ScaffoldDirs(projectPath string) error {
	var structure map[string]FileStructure
	err := json.Unmarshal(projectStructure, &structure)
	if err != nil {
		return fmt.Errorf("failed to parse structure file: %v", err)
	}

	// create project structure
	return createStructure(projectPath, structure)
}

// createStructure recursively creates directories and files
func createStructure(basePath string, structure map[string]FileStructure) error {
	for name, item := range structure {
		path := filepath.Join(basePath, name)

		if item.Type == "directory" {
			// create directory
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %v", path, err)
			}

			// recursively create children if they exist
			if item.Children != nil {
				err = createStructure(path, item.Children)
				if err != nil {
					return err
				}
			}
		} else if item.Type == "file" {
			// create empty file
			f, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", path, err)
			}
			f.Close()
		}
	}

	return nil
}
