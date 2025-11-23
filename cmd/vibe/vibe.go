package vibe

import (
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "vibe",
	Short: "Space for vibe-coded stuff",
}
