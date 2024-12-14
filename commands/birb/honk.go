// commands/birb/honk.go
package birb

import (
	"fmt"
	"strings"

	"github.com/rAlexander89/swan/nodes"
)

func init() {
	nodes.RegisterCommand("honk", Honk)
	nodes.RegisterCommand("says", Says)
}

func Honk([]string) error {
	fmt.Println("HONK! 🦢")
	return nil
}

func Says(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("swan needs something to say")
	}

	// join all arguments into a single message
	say := strings.Join(args, " ") + " - 🦢"
	fmt.Println(say)
	return nil
}
