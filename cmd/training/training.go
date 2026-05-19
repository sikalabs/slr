package training

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "training",
	Aliases: []string{"tr", "t"},
	Short:   "Space for Ondrej Sika's training related stuff",
}
