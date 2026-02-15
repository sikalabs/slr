package du_gitlab_tls_update

import (
	"log"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slu/utils/exec_utils"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "du-gitlab-tls-update",
	Short: "Update TLS certificates for DU GitLab from Kubernetes and restart nginx",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		err := exec_utils.ExecShOut(`
KUBECONFIG=/root/.kube/config &&
slr get-tls-from-kubernetes \
  --namespace gitlab-proxy \
  --name gitlab.du.gov.cz-tls \
  --file-cert /tls-cert.pem \
  --file-key /tls-priv.pem &&
gitlab-ctl restart nginx &&
echo "DU GitLab TLS update completed successfully" &&
slr os du-notification "⚠️ DU GitLab TLS update completed successfully"
`)
		if err != nil {
			log.Fatal(err)
		}
	},
}
