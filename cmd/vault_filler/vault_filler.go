package vault_filler

import (
	"fmt"
	"os"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "vault-filler",
	Short: "Deprecated: Use 'slu vault filler' instead",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		fmt.Println("###################################################################")
		fmt.Println("#                                                                 #")
		fmt.Println("#                                                                 #")
		fmt.Println("#   Dont use `slr vault-filler`. Use `slu vault filler` instead   #")
		fmt.Println("#                                                                 #")
		fmt.Println("#                                                                 #")
		fmt.Println("###################################################################")
		os.Exit(1)
	},
}
