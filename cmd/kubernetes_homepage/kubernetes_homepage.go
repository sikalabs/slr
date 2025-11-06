package kubernetes_homepage

import (
	"os"

	"github.com/sikalabs/sikalabs-kubernetes-homepage/pkg/server"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagName string

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "kubernetes-homepage",
	Short: "Run sikalabs/sikalabs-kubernetes-homepage",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {

		server.Server(server.ServerOptions{
			Port:    8000,
			Domain:  os.Getenv("DOMAIN"),
			Cluster: os.Getenv("CLUSTER"),
		})

	},
}
