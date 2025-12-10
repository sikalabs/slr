package kubernetes

import (
	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

func init() {
	training.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:     "kubernetes",
	Aliases: []string{"k8s", "k"},
	Short:   "Kubernetes related commands",
}
