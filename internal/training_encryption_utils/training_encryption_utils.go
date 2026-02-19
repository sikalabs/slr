package training_encryption_utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"golang.org/x/term"
)

func GetPasswordOrDie() string {
	passwordFilePath := os.Getenv("HOME") + "/.SLR_TRAINING_ENCRYPTION_PASSWORD"
	data, err := os.ReadFile(passwordFilePath)
	if err == nil {
		return strings.TrimRight(string(data), "\n")
	}
	return readPassword()
}

func readPassword() string {
	fmt.Fprint(os.Stderr, "Encryption Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	error_utils.HandleError(err)
	return string(password)
}
