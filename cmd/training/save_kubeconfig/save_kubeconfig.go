package save_kubeconfig

import (
	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

func init() {
	training.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:     "save-kubeconfig",
	Short:   "Save kubeconfig to Redis",
	Aliases: []string{"save-k"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		saveKubeconfigToRedis()
	},
}

func saveKubeconfigToRedis() {
	// TODO: Implement saving kubeconfig to Redis here
}
