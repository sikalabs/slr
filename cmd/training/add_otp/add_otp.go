package add_otp

import (
	"fmt"
	"os"

	"github.com/sikalabs/slr/cmd/training"
	"github.com/sikalabs/slr/internal/kv"
	"github.com/spf13/cobra"
)

func init() {
	training.Cmd.AddCommand(Cmd)
}

var FlagFile string

var Cmd = &cobra.Command{
	Use:   "add-otp <name> <secret>",
	Short: "Add OTP to key-value storage",
	Args:  cobra.ExactArgs(2),
	Run: func(c *cobra.Command, args []string) {
		err := addOTPToStorage(args[0], args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OTP added successfully")
	},
}

func addOTPToStorage(name string, secret string) error {
	key := "otp-" + name
	err := kv.Set(key, secret)
	if err != nil {
		return fmt.Errorf("failed to save OTP: %w", err)
	}
	return nil
}
