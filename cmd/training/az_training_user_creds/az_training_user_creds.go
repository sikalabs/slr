package az_training_user_creds

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/sikalabs/slr/cmd/training"
	"github.com/spf13/cobra"
)

const dataURL = "https://raw.githubusercontent.com/ondrejsika/training-cli-data/refs/heads/master/data/azure_training_user.json"

type AzureTrainingUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

func init() {
	training.Cmd.AddCommand(Cmd)
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

		code, err := totp.GenerateCode(user.OTP, time.Now())
		if err != nil {
			log.Fatal("Error generating OTP: ", err)
		}

		fmt.Printf("Username: %s\n", user.Username)
		fmt.Printf("Password: %s\n", user.Password)
		fmt.Printf("OTP:      %s-%s\n", code[:3], code[3:])
	},
}
