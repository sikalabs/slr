package kubernetes_oidc_login

import (
	"github.com/sikalabs/sikalabs-kubernetes-oidc-login/pkg/cmd"
	"github.com/sikalabs/slr/cmd/root"
)

func init() {
	Cmd := cmd.GetCmd(cmd.GetCmdOpts{
		NameOverride: "kubernetes-oidc-login",
	})
	root.Cmd.AddCommand(Cmd)
}
