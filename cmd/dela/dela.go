package dela

import (
	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slu/lib/printdela"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "dela",
	Short: "Show picture of Dela",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		printdela.PrintDela()
	},
}
