package cpu_load_generator

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagCores int

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().IntVarP(&FlagCores, "cores", "c", runtime.NumCPU(), "Number of CPU cores to use for load generation (default: all available cores)")
}

var Cmd = &cobra.Command{
	Use:   "cpu-load-generator",
	Short: "CPU Load Generator",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		cpuLoadGenerator(FlagCores)
	},
}

func cpuLoadGenerator(cores int) {
	runtime.GOMAXPROCS(cores)
	var wg sync.WaitGroup
	for i := 0; i < cores; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				_ = 123456789 * 987654321
			}
		}(&wg)
	}
	fmt.Printf("Generating CPU load with %d goroutines...\n", cores)
	wg.Wait()
}
