package set_encryption_password

import (
	"fmt"
	"os"

	"github.com/sikalabs/slr/cmd/training"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var FlagPassword string

func init() {
	training.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagPassword, "password", "p", "", "Encryption password")
}

var Cmd = &cobra.Command{
	Use:     "set-encryption-password",
	Aliases: []string{"sep"},
	Short:   "Save training encryption password to ~/.SLR_TRAINING_ENCRYPTION_PASSWORD",
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		setEncryptionPassword()
	},
}

func setEncryptionPassword() {
	var password []byte
	if FlagPassword != "" {
		password = []byte(FlagPassword)
	} else {
		fmt.Fprint(os.Stderr, "Encryption Password: ")
		var err error
		password, err = term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		error_utils.HandleError(err)
	}

	passwordFilePath := os.Getenv("HOME") + "/.SLR_TRAINING_ENCRYPTION_PASSWORD"
	err := os.WriteFile(passwordFilePath, append(password, '\n'), 0600)
	error_utils.HandleError(err)

	fmt.Println("Password saved to", passwordFilePath)
}
