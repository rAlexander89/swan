// main.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("error getting current directory:", err)
		os.Exit(1)
	}

	// get gopath using go env command instead of environment variable
	cmd := exec.Command("go", "env", "GOPATH")
	goPathBytes, err := cmd.Output()
	if err != nil {
		fmt.Println("error getting gopath:", err)
		os.Exit(1)
	}

	// trim whitespace and newlines from gopath
	goPath := strings.TrimSpace(string(goPathBytes))

	if goPath == "" {
		fmt.Println("gopath not set")
		os.Exit(1)
	}

	// check if pwd is within gopath
	if strings.HasPrefix(pwd, goPath) {
		fmt.Println("nice")
	} else {
		fmt.Printf("current directory (%s) is not in gopath (%s)\n", pwd, goPath)
		os.Exit(1)
	}
}
