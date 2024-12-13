// main.go
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("error getting current directory:", err)
		os.Exit(1)
	}

	// get gopath
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		fmt.Println("gopath not set")
		os.Exit(1)
	}

	// check if pwd is within gopath
	if strings.HasPrefix(pwd, goPath) {
		fmt.Println("nice")
	} else {
		fmt.Println("current directory is not in gopath")
		os.Exit(1)
	}
}
