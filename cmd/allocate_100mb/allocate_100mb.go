package allocate_100mb

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "allocate-100mb",
	Short: "Allocate 100MB of memory and wait for Ctrl+C",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		allocate100mb()
	},
}

func allocate100mb() {
	// Allocate 100 MB
	size := 100 * 1024 * 1024 // 100 MB
	data := make([]byte, size)

	// Use the memory to avoid optimization
	for i := range data {
		data[i] = 42
	}

	fmt.Println("Allocated 100MB of memory. Press Ctrl+C to exit.")

	// Wait for Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Exiting...")
}
