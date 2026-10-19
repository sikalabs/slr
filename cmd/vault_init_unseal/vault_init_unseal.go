package vault_init_unseal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagPath string
var FlagNamespace string
var FlagKeyShares int
var FlagKeyThreshold int

var Cmd = &cobra.Command{
	Use:     "vault-init-unseal",
	Short:   "Deprecated: Use 'slu vault init-unseal' instead",
	Args:    cobra.NoArgs,
	Aliases: []string{"viu"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`
####################################################################
#                                                                  #
#                                                                  #
#   Dont use ` + "`slr vault-init-unseal`. Use `slu vault init-unseal" + `   #
#                                                                  #
#                                                                  #
#                                                                  #
####################################################################`)
		run(true, "slu", "vault", "init-unseal",
			"--file", filepath.Join(FlagPath, "vault_keys.local.json"),
			"--namespace", FlagNamespace,
			"--key-shares", strconv.Itoa(FlagKeyShares),
			"--key-threshold", strconv.Itoa(FlagKeyThreshold),
		)
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.MarkFlagRequired("path")
	Cmd.Flags().StringVarP(
		&FlagPath,
		"path",
		"p",
		"",
		"Path to save vault keys",
	)
	Cmd.Flags().StringVarP(
		&FlagNamespace,
		"namespace",
		"n",
		"vault",
		"Namespace where Vault is running",
	)
	Cmd.Flags().IntVarP(
		&FlagKeyShares,
		"key-shares",
		"s",
		5,
		"Number of key shares to generate during Vault initialization",
	)
	Cmd.Flags().IntVarP(
		&FlagKeyThreshold,
		"key-threshold",
		"t",
		3,
		"Number of key shares required to unseal Vault",
	)
}

func run(failOnError bool, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil && failOnError {
		log.Fatalf("Error running %s: %v", name, err)
	}
}
