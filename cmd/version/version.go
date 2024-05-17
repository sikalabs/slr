package version

import (
	"fmt"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/version"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "version",
	Short:   "Prints version",
	Aliases: []string{"v"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		fmt.Printf("%s\n", version.Version)
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
}
