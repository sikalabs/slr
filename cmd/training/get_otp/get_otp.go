package get_otp

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/sikalabs/slr/cmd/training"
	"github.com/sikalabs/slr/internal/kv"
	"github.com/spf13/cobra"
)

func init() {
	training.Cmd.AddCommand(Cmd)
}

var FlagFile string

var Cmd = &cobra.Command{
	Use:   "get-otp <name>",
	Short: "Get OTP from key-value storage",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		secret, err := getOTPFromStorage(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		now := time.Now()
		code, err := totp.GenerateCode(secret, now)
		if err != nil {
			log.Printf("Error generating OTP for %s: %v", args[0], err)
			os.Exit(1)
		}
		fmt.Println("OTP:", code)
	},
}

func getOTPFromStorage(name string) (string, error) {
	key := "otp-" + name
	otp, err := kv.Get(key)
	if err != nil {
		return "", fmt.Errorf("failed to get OTP: %w", err)
	}
	return otp, nil
}
