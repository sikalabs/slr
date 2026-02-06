package save_env_to_file

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "save-env-to-file",
	Short: "Save current environment variables to a file",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf(".saved_env.%s", timestamp)

		envs := os.Environ()
		sort.Strings(envs)

		err := os.WriteFile(filename, []byte(strings.Join(envs, "\n")+"\n"), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
			os.Exit(1)
		}

		absPath, err := filepath.Abs(filename)
		if err != nil {
			absPath = filename
		}
		fmt.Printf("Environment saved to %s\n", absPath)
	},
}
