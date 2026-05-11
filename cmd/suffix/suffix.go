package suffix

import (
	"fmt"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "suffix",
	Short: "Print date suffix in YYMMDD format",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		fmt.Println(time.Now().Format("060102") + " (YYMMDD format)")
	},
}
