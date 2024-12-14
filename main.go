// main.go
package main

import (
	"fmt"
	"os"

	_ "github.com/rAlexander89/swan/commands/birb"
	_ "github.com/rAlexander89/swan/commands/project"
	"github.com/rAlexander89/swan/nodes"
)

type argument struct {
	key       string
	value     string
	fields    []field
	modifiers map[string]string // stores --flag=value pairs
}

type field struct {
	name     string
	dataType string
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("no command provided")
		os.Exit(1)
	}

	// load node tree
	root, err := nodes.LoadNodes()
	if err != nil {
		fmt.Printf("error loading nodes: %v\n", err)
		os.Exit(1)
	}

	// get the command from args
	cmd := args[0]
	remainingArgs := args[1:]

	// find node and get registered function
	node, exists := root.BranchMap[cmd]
	if !exists {
		fmt.Printf("unknown command: %s\n", cmd)
		os.Exit(1)
	}

	fn, fnExists := nodes.GetCommand(cmd)
	if !fnExists {
		fmt.Printf("command %s not implemented\n", cmd)
		os.Exit(1)
	}
	node.Run = fn

	if err := node.Run(remainingArgs); err != nil {
		fmt.Printf("error executing command: %v\n", err)
		os.Exit(1)
	}
}
