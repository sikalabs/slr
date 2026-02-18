package az_training_user_creds

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/sikalabs/sikalabs-crypt-go/pkg/sikalabs_crypt"
	"github.com/sikalabs/slr/cmd/training"
	"github.com/sikalabs/slr/internal/training_encryption_utils"
	"github.com/sikalabs/slu/pkg/utils/error_utils"
	"github.com/spf13/cobra"
)

const dataURL = "https://raw.githubusercontent.com/ondrejsika/training-cli-data/refs/heads/master/data/azure_training_user.json"

var FlagPassword string

type AzureTrainingUser struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	PasswordEncrypted string `json:"password_encrypted"`
	OTP               string `json:"otp"`
	OTPEncrypted      string `json:"otp_encrypted"`
}

func init() {
	training.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagPassword, "password", "p", "", "Decryption password (prompted if not set)")
}

var Cmd = &cobra.Command{
	Use:   "az-training-user-creds",
	Short: "Get Azure training user credentials",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		resp, err := http.Get(dataURL)
		if err != nil {
			log.Fatal("Error fetching data: ", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading response: ", err)
		}

		var user AzureTrainingUser
		if err := json.Unmarshal(body, &user); err != nil {
			log.Fatal("Error parsing JSON: ", err)
		}

		needsDecryption := (user.Password == "" && user.PasswordEncrypted != "") ||
			(user.OTP == "" && user.OTPEncrypted != "")

		password := FlagPassword
		if needsDecryption && password == "" {
			password = training_encryption_utils.GetPasswordOrDie()
		}

		if user.Password == "" && user.PasswordEncrypted != "" {
			user.Password = decrypt(user.PasswordEncrypted, password)
		}
		if user.OTP == "" && user.OTPEncrypted != "" {
			user.OTP = decrypt(user.OTPEncrypted, password)
		}

		code, err := totp.GenerateCode(user.OTP, time.Now())
		if err != nil {
			log.Fatal("Error generating OTP: ", err)
		}

		fmt.Printf("Username: %s\n", user.Username)
		fmt.Printf("Password: %s\n", user.Password)
		fmt.Printf("OTP:      %s-%s\n", code[:3], code[3:])
	},
}

func decrypt(encrypted, password string) string {
	decrypted, err := sikalabs_crypt.SikaLabsSymmetricDecryptV1(password, encrypted)
	error_utils.HandleError(err)
	return decrypted
}
