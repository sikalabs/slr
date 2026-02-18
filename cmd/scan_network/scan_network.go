package scan_network

import (
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"sync"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "scan-network <cidr>",
	Short: "Scan all IPs in a network range (e.g. 192.168.1.0/24)",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		scanNetwork(args[0])
	},
}

func scanNetwork(cidr string) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Printf("Error: invalid CIDR %q: %v\n", cidr, err)
		return
	}

	ips := allIPs(ip, ipNet)
	fmt.Printf("Scanning %d hosts in %s...\n", len(ips), cidr)

	var mu sync.Mutex
	var wg sync.WaitGroup
	var up []string

	for _, host := range ips {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if ping(host) {
				mu.Lock()
				up = append(up, host)
				mu.Unlock()
			}
		}(host)
	}

	wg.Wait()

	if len(up) == 0 {
		fmt.Println("No hosts responded.")
		return
	}

	fmt.Printf("\nUp (%d):\n", len(up))
	for _, h := range up {
		fmt.Println(" ", h)
	}
}

func allIPs(base net.IP, ipNet *net.IPNet) []string {
	// Work with 4-byte representation
	ip4 := base.To4()
	if ip4 == nil {
		fmt.Println("Error: only IPv4 networks are supported")
		return nil
	}
	mask := ipNet.Mask
	network := binary.BigEndian.Uint32([]byte(ipNet.IP.To4()))
	broadcast := network | ^binary.BigEndian.Uint32([]byte(mask))

	var ips []string
	// Skip network address (network) and broadcast address (broadcast)
	for n := network + 1; n < broadcast; n++ {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, n)
		ips = append(ips, net.IP(b).String())
	}
	return ips
}

func ping(host string) bool {
	err := exec.Command("ping", "-c", "1", "-W", "1", host).Run()
	return err == nil
}
