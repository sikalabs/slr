package root

import (
	"github.com/sikalabs/slr/version"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "slr",
	Short: "slr, " + version.Version,
}
