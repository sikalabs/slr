package cmd

import (
	_ "github.com/sikalabs/slr/cmd/redis_set_large_data"
	"github.com/sikalabs/slr/cmd/root"
	_ "github.com/sikalabs/slr/cmd/test_clisso_cli"
	_ "github.com/sikalabs/slr/cmd/version"
	"github.com/spf13/cobra"
)

func Execute() {
	cobra.CheckErr(root.Cmd.Execute())
}
