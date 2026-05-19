package azure

import (
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:     "azure",
	Aliases: []string{"az"},
	Short:   "Azure Utilities",
}
