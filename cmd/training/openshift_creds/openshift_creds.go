package openshift_creds

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/sikalabs/sikalabs-crypt-go/pkg/sikalabs_crypt"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"golang.org/x/term"

	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

const dataURL = "https://raw.githubusercontent.com/ondrejsika/training-cli-data/refs/heads/master/data/openshift_creds.json"

var FlagAll bool
var FlagPassword string

type OpenShiftCreds struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	PasswordEncrypted string `json:"password_encrypted"`
	ClusterName       string `json:"cluster_name"`
	ConsoleURL        string `json:"console_url"`
}

func init() {
	training.Cmd.AddCommand(Cmd)
	Cmd.Flags().BoolVarP(&FlagAll, "all", "a", false, "Show all credentials")
	Cmd.Flags().StringVarP(&FlagPassword, "password", "p", "", "Decryption password (prompted if not set)")
}

var Cmd = &cobra.Command{
	Use:   "openshift-creds",
	Short: "Get OpenShift training credentials",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		password := FlagPassword
		if password == "" {
			password = readPassword()
		}

		resp, err := http.Get(dataURL)
		if err != nil {
			log.Fatal("Error fetching data: ", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading response: ", err)
		}

		var creds []OpenShiftCreds
		if err := json.Unmarshal(body, &creds); err != nil {
			log.Fatal("Error parsing JSON: ", err)
		}

		if len(creds) == 0 {
			log.Fatal("No credentials found")
		}

		if FlagAll {
			for i, cred := range creds {
				if i > 0 {
					fmt.Println()
				}
				printCreds(cred, password)
			}
		} else {
			printCreds(creds[0], password)
		}
	},
}

func readPassword() string {
	fmt.Fprint(os.Stderr, "Encryption Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	error_utils.HandleError(err)
	return string(password)
}

func decryptPassword(encrypted, password string) string {
	decrypted, err := sikalabs_crypt.SikaLabsSymmetricDecryptV1(password, encrypted)
	error_utils.HandleError(err)
	return decrypted
}

func printCreds(cred OpenShiftCreds, encryptionPassword string) {
	password := cred.Password
	if password == "" && cred.PasswordEncrypted != "" {
		password = decryptPassword(cred.PasswordEncrypted, encryptionPassword)
	}

	fmt.Printf("Cluster:  %s\n", cred.ClusterName)
	if cred.ConsoleURL != "" {
		fmt.Printf("Console:  %s\n", cred.ConsoleURL)
	}
	fmt.Printf("Username: %s\n", cred.Username)
	fmt.Printf("Password: %s\n", password)
}
