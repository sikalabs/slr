package validate_jwt

import (
	"context"
	"log"

	"github.com/coreos/go-oidc"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "validate-jwt <issuer> <rawToken>",
	Short: "Validate JWT",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		validateJTW(args[0], args[1])
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
}

func validateJTW(issuer, rawToken string) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatal(err)
	}

	_, err = provider.Verifier(&oidc.Config{SkipClientIDCheck: true}).Verify(ctx, rawToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("JWT Token is valid")
}
