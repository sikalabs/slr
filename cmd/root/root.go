package root

import (
	_ "github.com/sikalabs/install-slu/cmd"
	install_slu_root "github.com/sikalabs/install-slu/cmd/root"
	"github.com/sikalabs/slr/version"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "slr",
	Short: "slr, " + version.Version,
}

func init() {
	Cmd.AddCommand(install_slu_root.Cmd)
}
