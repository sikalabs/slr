package memory_usage

import (
	"fmt"
	"log"
	"strconv"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.MarkFlagRequired("pid")
}

var Cmd = &cobra.Command{
	Use:   "memory-usage <pid>",
	Short: "Get memory usage of a process",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Error converting PID to int: %v\n", err)
		}
		memoryUsage(pid)
	},
}

func memoryUsage(pid int) {

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Fatalf("Error getting process: %v\n", err)
	}

	memInfo, err := p.MemoryInfo()
	if err != nil {
		log.Fatalf("Error getting memory info: %v\n", err)
	}

	fmt.Printf("PID: %d\n", pid)
	fmt.Printf("RSS: %v bytes (%.2f MB)\n", memInfo.RSS, float64(memInfo.RSS)/(1024*1024))
	fmt.Printf("VMS: %v bytes (%.2f MB)\n", memInfo.VMS, float64(memInfo.VMS)/(1024*1024))
}
