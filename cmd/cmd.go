package cmd

import (
	"github.com/sikalabs/slr/cmd/root"
	_ "github.com/sikalabs/slr/cmd/version"
	"github.com/spf13/cobra"
)

func Execute() {
	cobra.CheckErr(root.Cmd.Execute())
}
