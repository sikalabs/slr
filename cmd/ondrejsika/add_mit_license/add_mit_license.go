package add_mit_license

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "add-mit-license",
	Short: "Deprecated: Use 'slu os add-mit-license' instead",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		fmt.Println(`######################################################################
#                                                                    #
#                                                                    #
#   Dont use ` + "`slr os add-mit-license`. Use `slu os add-mit-license`" + `  #
#                                                                    #
#                                                                    #
#                                                                    #
######################################################################`)
		run("slu", "os", "add-mit-license")
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
