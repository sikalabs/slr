package otp

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "otp <secret>",
	Short: "Deprecated: Use 'slu s otp <secret>' instead.",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		fmt.Println(`######################################################################
#                                                                    #
#                                                                    #
#   Dont use ` + "`slr otp <secret>`. Use `slu s otp <secret>`" + `            #
#   Or use ` + "`slu scripts otp <secret>`" + `                                #
#                                                                    #
#                                                                    #
######################################################################`)
		fmt.Println()
		run("slu", "s", "otp", args[0])
	},
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error running %s: %v", name, err)
	}
}
