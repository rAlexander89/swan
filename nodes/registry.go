// nodes/registry.go
package nodes

var commandRegistry = make(map[string]func([]string) error)

func RegisterCommand(name string, fn func([]string) error) {
	commandRegistry[name] = fn
}

func GetCommand(name string) (func([]string) error, bool) {
	fn, exists := commandRegistry[name]
	return fn, exists
}
