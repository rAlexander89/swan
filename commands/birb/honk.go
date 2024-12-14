// commands/birb/honk.go
package birb

import (
	"fmt"

	"github.com/rAlexander89/swan/nodes"
)

func init() {
	nodes.RegisterCommand("honk", Honk)
}

func Honk([]string) error {
	fmt.Println("HONK! 🦢")
	return nil
}
