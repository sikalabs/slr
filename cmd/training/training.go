package training

import (
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:     "training",
	Aliases: []string{"tr", "t"},
	Short:   "Space for Ondrej Sika's training related stuff",
}
