package openshift_pull_secret

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sikalabs/sikalabs-crypt-go/pkg/sikalabs_crypt"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"golang.org/x/term"

	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

const openShiftPullSecretURL = "https://raw.githubusercontent.com/ondrejsika/training-cli-data/refs/heads/master/data/openshift_pull_secret.encrypted.txt"

var FlagPassword string

func init() {
	training.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagPassword,
		"password",
		"p",
		"",
		"Decryption password (prompted if not set)",
	)
}

var Cmd = &cobra.Command{
	Use:   "openshift-pull-secret",
	Short: "Get OpenShift pull secret",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		encrypted := wgetOrDie(openShiftPullSecretURL)
		decrypted := decryptOrDie(encrypted)
		fmt.Println(decrypted)
	},
}

func readPassword() string {
	fmt.Fprint(os.Stderr, "Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	error_utils.HandleError(err)
	return string(password)
}

func decryptOrDie(encrypted string) string {
	password := FlagPassword
	if password == "" {
		password = readPassword()
	}
	decrypted, err := sikalabs_crypt.SikaLabsSymmetricDecryptV1(
		password,
		encrypted,
	)
	error_utils.HandleError(err)
	return decrypted
}

func wgetOrDie(url string) string {
	resp, err := http.Get(url)
	error_utils.HandleError(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	error_utils.HandleError(err)
	return string(body)
}
