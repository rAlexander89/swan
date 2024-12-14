// main.go
package main

import (
	"fmt"
	"os"

	_ "github.com/rAlexander89/swan/commands/birb"
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

	// find node and get registered function
	if node, exists := root.BranchMap[cmd]; exists {
		if fn, exists := nodes.GetCommand(cmd); exists {
			node.Run = fn
			if err := node.Run(args[1:]); err != nil {
				fmt.Printf("error executing command: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("command %s not implemented\n", cmd)
			os.Exit(1)
		}
	} else {
		fmt.Printf("unknown command: %s\n", cmd)
		os.Exit(1)
	}
}
