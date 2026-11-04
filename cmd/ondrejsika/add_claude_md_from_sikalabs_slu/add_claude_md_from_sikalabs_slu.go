package add_claude_md_from_sikalabs_slu

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
	Use:     "add-claude-md-from-sikalabs-slu",
	Short:   "Deprecated: Use 'slu os add-claude-md-from-sikalabs-slu' instead",
	Aliases: []string{"add-claude-md"},
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		fmt.Println(`##############################################################################
#                                                                            #
#                                                                            #
#   Dont use ` + "`slr os add-claude-md-from-sikalabs-slu`" + `.                       #
#   Use ` + "`slu os add-claude-md-from-sikalabs-slu`" + `                             #
#                                                                            #
#                                                                            #
##############################################################################`)
		run("slu", "os", "add-claude-md-from-sikalabs-slu")
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
