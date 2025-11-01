package stegosay

import (
	"fmt"
	"strings"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "stegosay <text>",
	Short: "Like cowsay but with stegosaurus",
	Args:  cobra.MinimumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		stegosay(strings.Join(args, " "))
	},
}

func stegosay(text string) {
	fmt.Print(bubble(text))
	fmt.Print(stegosaurus)
}

func bubble(text string) string {
	lines := strings.Split(text, "\n")
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	top := " " + strings.Repeat("_", maxLen+2)
	bottom := " " + strings.Repeat("-", maxLen+2)
	var middle []string
	for _, line := range lines {
		padded := line + strings.Repeat(" ", maxLen-len(line))
		middle = append(middle, fmt.Sprintf("< %s >", padded))
	}

	return fmt.Sprintf("%s\n%s\n%s\n", top, strings.Join(middle, "\n"), bottom)
}
