package prvninakup_notification

import (
	"fmt"
	"strings"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "prvninakup-notification <message>",
	Short: "Send a prvninakup notification to Ondrej Sika",
	Args:  cobra.MinimumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		prvninakupNotification(strings.Join(args, " "))
	},
}

func prvninakupNotification(message string) {
	fmt.Println(message)
}
