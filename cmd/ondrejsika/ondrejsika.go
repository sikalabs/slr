package ondrejsika

import (
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:     "ondrejsika",
	Aliases: []string{"os"},
	Short:   "Ondrej Sika's Random Utils",
	Args:    cobra.NoArgs,
}
