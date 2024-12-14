// main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

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
	if err := validateEnvironment(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	root, err := nodes.LoadNodes()
	if err != nil {
		fmt.Printf("error loading nodes: %v\n", err)
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("usage: swan <command> [args...]")
		os.Exit(1)
	}

	currentNode := root
	cmdIndex := 0

	for cmdIndex < len(args) {
		arg := args[cmdIndex]

		// handle primary command with modifiers
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			parsedArg, consumed := parseCommandWithModifiers(args[cmdIndex:])
			if err := executeWithModifiers(currentNode, parsedArg); err != nil {
				fmt.Printf("error executing command: %v\n", err)
				os.Exit(1)
			}
			cmdIndex += consumed
			continue
		}

		// regular command navigation
		nextNode, exists := currentNode.BranchMap[arg]
		if !exists {
			fmt.Printf("unknown command '%s'\n", arg)
			os.Exit(1)
		}

		currentNode = nextNode
		cmdIndex++

		if currentNode.Config != nil {
			remainingArgs := args[cmdIndex:]
			if err := executeCommand(currentNode.Config, remainingArgs); err != nil {
				fmt.Printf("error executing command: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func parseCommandWithModifiers(args []string) (argument, int) {
	arg := argument{
		modifiers: make(map[string]string),
	}
	argsConsumed := 0

	// parse primary command (e.g., -User=...)
	primary := strings.TrimPrefix(args[0], "-")
	parts := strings.SplitN(primary, "=", 2)
	arg.key = parts[0]

	if len(parts) > 1 {
		// parse fields for struct-like commands
		fieldStr := parts[1]
		fieldParts := strings.Split(fieldStr, ",")

		for _, f := range fieldParts {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}

			parts := strings.Fields(f)
			if len(parts) >= 2 {
				arg.fields = append(arg.fields, field{
					name:     parts[0],
					dataType: parts[1],
				})
			}
		}
	}
	argsConsumed++

	// look for modifiers (--flag=value)
	for i := 1; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			break
		}

		modifier := strings.TrimPrefix(args[i], "--")
		modParts := strings.SplitN(modifier, "=", 2)

		if len(modParts) == 2 {
			arg.modifiers[modParts[0]] = modParts[1]
		} else {
			arg.modifiers[modParts[0]] = "true"
		}

		argsConsumed++
	}

	return arg, argsConsumed
}

func executeWithModifiers(node *nodes.Node, arg argument) error {
	// example handling for struct generation
	if len(arg.fields) > 0 {
		fmt.Printf("generating struct %s with fields:\n", arg.key)
		for _, field := range arg.fields {
			fmt.Printf("  %s %s\n", field.name, field.dataType)
		}

		// print modifiers
		for key, value := range arg.modifiers {
			fmt.Printf("modifier: --%s=%s\n", key, value)
		}

		// handle specific modifiers
		if tags, ok := arg.modifiers["tags"]; ok {
			fmt.Printf("will add tags: %s\n", tags)
		}
	} else {
		// handle simple commands with modifiers
		fmt.Printf("executing command %s with modifiers:\n", arg.key)
		for key, value := range arg.modifiers {
			fmt.Printf("  --%s=%s\n", key, value)
		}
	}

	return nil
}

func validateEnvironment() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return fmt.Errorf("gopath not set")
	}

	if !strings.HasPrefix(pwd, goPath) {
		return fmt.Errorf("current directory is not in gopath")
	}

	return nil
}

func executeCommand(config *nodes.Config, args []string) error {
	// validate args if specified in config
	if config.Args != nil {
		if err := validateArgs(config.Args, args); err != nil {
			return fmt.Errorf("invalid arguments: %v", err)
		}
	}

	// get the package path relative to the project root
	pkgPath := filepath.Join("github.com/rAlexander89/swan", config.Package)

	di := newDynamicImporter()

	// import the package
	pkg, err := di.importPackage(pkgPath)
	if err != nil {
		return fmt.Errorf("error importing package %s: %v", pkgPath, err)
	}

	// get the method
	method := pkg.MethodByName(config.Function)
	if !method.IsValid() {
		return fmt.Errorf("function %s not found in package %s",
			config.Function, config.Package)
	}

	// prepare arguments based on the function signature
	fnType := method.Type()
	var fnArgs []reflect.Value

	// if function accepts arguments and we have args to pass
	if fnType.NumIn() > 0 && len(args) > 0 {
		// convert args based on config.Args if present
		if config.Args != nil {
			convertedArgs, err := convertArgs(config.Args, args)
			if err != nil {
				return fmt.Errorf("error converting arguments: %v", err)
			}
			fnArgs = convertedArgs
		} else {
			// if no arg config, pass raw string slice
			fnArgs = append(fnArgs, reflect.ValueOf(args))
		}
	}

	// call the function
	results := method.Call(fnArgs)

	// check for error return
	if len(results) > 0 {
		last := results[len(results)-1].Interface()
		if err, ok := last.(error); ok && err != nil {
			return err
		}
	}

	return nil
}

func validateArgs(configArgs *[]struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}, providedArgs []string,
) error {
	if configArgs == nil {
		return nil
	}

	// check required args
	requiredCount := 0
	for _, arg := range *configArgs {
		if arg.Required {
			requiredCount++
		}
	}

	if len(providedArgs) < requiredCount {
		return fmt.Errorf("not enough arguments provided. expected at least %d, got %d",
			requiredCount, len(providedArgs))
	}

	return nil
}

func convertArgs(configArgs *[]struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}, providedArgs []string,
) ([]reflect.Value, error) {
	var converted []reflect.Value

	if configArgs == nil {
		return converted, nil
	}

	for i, arg := range *configArgs {
		if i >= len(providedArgs) {
			if arg.Required {
				return nil, fmt.Errorf("missing required argument: %s", arg.Name)
			}
			continue
		}

		val, err := convertArgument(providedArgs[i], arg.Type)
		if err != nil {
			return nil, fmt.Errorf("error converting argument %s: %v", arg.Name, err)
		}
		converted = append(converted, val)
	}

	return converted, nil
}

func convertArgument(value string, targetType string) (reflect.Value, error) {
	switch targetType {
	case "string":
		return reflect.ValueOf(value), nil
	case "int":
		v, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v), nil
	case "[]string":
		parts := strings.Split(value, ",")
		return reflect.ValueOf(parts), nil
	// add more types as needed
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type: %s", targetType)
	}
}
