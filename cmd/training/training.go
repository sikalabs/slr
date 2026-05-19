package training

import (
	"github.com/sikalabs/slr/version"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "training",
	Aliases: []string{"tr", "t"},
	Short:   "Space for Ondrej Sika's training related stuff (slr " + version.Version + ")",
}
