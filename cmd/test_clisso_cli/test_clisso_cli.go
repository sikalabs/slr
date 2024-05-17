package test_clisso_cli

import (
	"fmt"
	"log"

	"github.com/mlosinsky/clisso/ssoclient"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

const DEFAULT_LOGIN_URL = "https://test-clisso-proxy.aks-prod.k8s.sl.zone/cli-login"

var FlagLoginUrl string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagLoginUrl,
		"login-url",
		"u",
		DEFAULT_LOGIN_URL,
		"Login URL",
	)
}

var Cmd = &cobra.Command{
	Use:   "test-clisso-cli",
	Short: "test mlosinsy/clisso CLI",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		clissoCliTest(FlagLoginUrl)
	},
}

func clissoCliTest(loginUrl string) {
	if loginResult, err := ssoclient.LoginWithSSOProxy(
		loginUrl,
		func(loginURL string) {
			fmt.Println("Login at:", loginURL)
		},
	); err != nil {
		log.Fatalln(err)
	} else {
		fmt.Printf("%+v\n", loginResult)
	}
}
