package get_hw_info

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "get-hw-info",
	Short: "Print hardware info (CPU, memory, disks, graphics cards)",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		getHwInfo()
	},
}

func getHwInfo() {
	printHostInfo()
	printCpuInfo()
	printMemoryInfo()
	printDiskInfo()
	printGpuInfo()
}

func printHostInfo() {
	fmt.Println("== Host ==")
	info, err := host.Info()
	if err != nil {
		log.Printf("Error getting host info: %v\n", err)
		return
	}
	fmt.Printf("Hostname: %s\n", info.Hostname)
	fmt.Printf("OS: %s (%s %s)\n", info.OS, info.Platform, info.PlatformVersion)
	fmt.Printf("Kernel: %s %s\n", info.KernelVersion, info.KernelArch)
	fmt.Println()
}

func printCpuInfo() {
	fmt.Println("== CPU ==")
	infos, err := cpu.Info()
	if err != nil {
		log.Printf("Error getting CPU info: %v\n", err)
		return
	}
	for i, ci := range infos {
		fmt.Printf("CPU %d: %s (%.0f MHz)\n", i, ci.ModelName, ci.Mhz)
	}
	physicalCount, _ := cpu.Counts(false)
	logicalCount, _ := cpu.Counts(true)
	fmt.Printf("Physical cores: %d\n", physicalCount)
	fmt.Printf("Logical cores: %d\n", logicalCount)
	fmt.Println()
}

func printMemoryInfo() {
	fmt.Println("== Memory ==")
	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory info: %v\n", err)
		return
	}
	fmt.Printf("Total: %.2f GB\n", bytesToGB(vm.Total))
	fmt.Printf("Used: %.2f GB (%.1f%%)\n", bytesToGB(vm.Used), vm.UsedPercent)
	fmt.Printf("Available: %.2f GB\n", bytesToGB(vm.Available))
	fmt.Println()
}

func printDiskInfo() {
	fmt.Println("== Disks ==")
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Printf("Error getting disk partitions: %v\n", err)
		return
	}
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		fmt.Printf("%s (%s, %s)\n", p.Mountpoint, p.Device, p.Fstype)
		fmt.Printf("  Total: %.2f GB, Used: %.2f GB (%.1f%%), Free: %.2f GB\n",
			bytesToGB(usage.Total), bytesToGB(usage.Used), usage.UsedPercent, bytesToGB(usage.Free))
	}
	fmt.Println()
}

func printGpuInfo() {
	fmt.Println("== Graphics Cards ==")
	gpus, err := getGpuInfo()
	if err != nil {
		log.Printf("Error getting GPU info: %v\n", err)
		return
	}
	if len(gpus) == 0 {
		fmt.Println("No graphics cards found")
		return
	}
	for _, g := range gpus {
		fmt.Println(g)
	}
}

func bytesToGB(b uint64) float64 {
	return float64(b) / (1024 * 1024 * 1024)
}

func getGpuInfo() ([]string, error) {
	switch runtime.GOOS {
	case "linux":
		return getGpuInfoLinux()
	case "darwin":
		return getGpuInfoDarwin()
	case "windows":
		return getGpuInfoWindows()
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func getGpuInfoLinux() ([]string, error) {
	out, err := exec.Command("lspci").Output()
	if err != nil {
		return nil, err
	}
	var gpus []string
	for _, line := range strings.Split(string(out), "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "vga compatible controller") || strings.Contains(lower, "3d controller") {
			gpus = append(gpus, line)
		}
	}
	return gpus, nil
}

func getGpuInfoDarwin() ([]string, error) {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return nil, err
	}
	var gpus []string
	for _, line := range strings.Split(string(out), "\n") {
		trimmed := strings.TrimSpace(line)
		if name, ok := strings.CutPrefix(trimmed, "Chipset Model:"); ok {
			gpus = append(gpus, strings.TrimSpace(name))
		}
	}
	return gpus, nil
}

func getGpuInfoWindows() ([]string, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-CimInstance Win32_VideoController | ForEach-Object { $_.Name }").Output()
	if err != nil {
		return nil, err
	}
	var gpus []string
	for _, line := range strings.Split(string(out), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			gpus = append(gpus, trimmed)
		}
	}
	return gpus, nil
}
