package gojekyll

import (
	"fmt"
	"os"

	"github.com/osteele/gojekyll/commands"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "gojekyll [...]",
	Short: "Embed osteele/gojekyll",
	Run: func(cmd *cobra.Command, args []string) {
		err := commands.ParseAndRun(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
}
