package check_before_rke2_install

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagRole string
var FlagAirgap bool

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagRole, "role", "r", "server", "Node role: server or agent")
	Cmd.Flags().BoolVar(&FlagAirgap, "airgap", false, "Skip internet connectivity checks (airgap installation)")
}

var Cmd = &cobra.Command{
	Use:   "check-before-rke2-install",
	Short: "Check if this server meets all requirements for RKE2 node installation",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if FlagRole != "server" && FlagRole != "agent" {
			fmt.Fprintln(os.Stderr, "--role must be server or agent")
			os.Exit(1)
		}
		if !checkBeforeRke2Install(FlagRole, FlagAirgap) {
			os.Exit(1)
		}
	},
}

type status string

const (
	statusOK   status = "OK"
	statusWarn status = "WARN"
	statusFail status = "FAIL"
	statusInfo status = "INFO"
)

type checker struct {
	fails int
	warns int
}

func (c *checker) report(s status, name, msg string) {
	switch s {
	case statusFail:
		c.fails++
	case statusWarn:
		c.warns++
	}
	fmt.Printf("[%-4s] %-28s %s\n", s, name, msg)
}

func checkBeforeRke2Install(role string, airgap bool) bool {
	c := &checker{}

	fmt.Printf("Checking requirements for RKE2 %s node ...\n\n", role)

	checkOS(c)
	if runtime.GOOS != "linux" {
		return false
	}
	checkRoot(c)
	checkCPU(c, role)
	checkMemory(c, role)
	checkDisk(c)
	checkSwap(c)
	checkHostname(c)
	checkCgroups(c)
	checkKernelModules(c)
	checkIPForward(c)
	checkSELinux(c)
	checkExistingInstallations(c)
	checkConflictingServices(c)
	checkNetworkManager(c)
	checkTimeSync(c)
	checkBinaries(c)
	checkDefaultRoute(c)
	checkPorts(c, role)
	if airgap {
		c.report(statusInfo, "Internet connectivity", "skipped (--airgap)")
	} else {
		checkInternet(c)
	}

	fmt.Printf("\nSummary: %d failed, %d warnings\n", c.fails, c.warns)
	if c.fails > 0 {
		fmt.Println("Server is NOT ready for RKE2 installation.")
		return false
	}
	fmt.Println("Server is ready for RKE2 installation.")
	return true
}

func checkOS(c *checker) {
	if runtime.GOOS != "linux" {
		c.report(statusFail, "Operating system", fmt.Sprintf("%s is not supported, RKE2 requires Linux", runtime.GOOS))
		return
	}
	name := runtime.GOOS
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				name = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
		}
	}
	c.report(statusOK, "Operating system", name)

	switch runtime.GOARCH {
	case "amd64", "arm64":
		c.report(statusOK, "Architecture", runtime.GOARCH)
	default:
		c.report(statusFail, "Architecture", runtime.GOARCH+" is not supported (amd64, arm64)")
	}
}

func checkRoot(c *checker) {
	if os.Geteuid() == 0 {
		c.report(statusOK, "Root privileges", "running as root")
	} else {
		c.report(statusFail, "Root privileges", "RKE2 installation requires root, run with sudo")
	}
}

func checkCPU(c *checker, role string) {
	n := runtime.NumCPU()
	min, rec := 2, 4
	if role == "agent" {
		min, rec = 1, 2
	}
	switch {
	case n < min:
		c.report(statusFail, "CPU", fmt.Sprintf("%d cores, minimum is %d", n, min))
	case n < rec:
		c.report(statusWarn, "CPU", fmt.Sprintf("%d cores, recommended is %d", n, rec))
	default:
		c.report(statusOK, "CPU", fmt.Sprintf("%d cores", n))
	}
}

func readMeminfo() map[string]uint64 {
	out := map[string]uint64{}
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return out
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) < 2 {
			continue
		}
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		out[strings.TrimSuffix(fields[0], ":")] = v * 1024
	}
	return out
}

func checkMemory(c *checker, role string) {
	mem := readMeminfo()
	total, ok := mem["MemTotal"]
	if !ok {
		c.report(statusWarn, "Memory", "cannot read /proc/meminfo")
		return
	}
	gb := float64(total) / (1 << 30)
	// small tolerance, MemTotal is always a bit lower than physical RAM
	min, rec := 3.5, 7.5
	minS, recS := "4GB", "8GB"
	if role == "agent" {
		min, rec = 1.5, 3.5
		minS, recS = "2GB", "4GB"
	}
	switch {
	case gb < min:
		c.report(statusFail, "Memory", fmt.Sprintf("%.1fGB, minimum is %s", gb, minS))
	case gb < rec:
		c.report(statusWarn, "Memory", fmt.Sprintf("%.1fGB, recommended is %s", gb, recS))
	default:
		c.report(statusOK, "Memory", fmt.Sprintf("%.1fGB", gb))
	}
}

func checkDisk(c *checker) {
	// /var/lib/rancher may not exist yet, check the closest existing parent
	path := "/var/lib/rancher"
	for {
		if _, err := os.Stat(path); err == nil {
			break
		}
		i := strings.LastIndex(path, "/")
		if i <= 0 {
			path = "/"
			break
		}
		path = path[:i]
	}
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		c.report(statusWarn, "Disk space", fmt.Sprintf("cannot stat %s: %s", path, err))
		return
	}
	gb := float64(uint64(st.Bavail)*uint64(st.Bsize)) / (1 << 30)
	msg := fmt.Sprintf("%.1fGB free on %s (/var/lib/rancher)", gb, path)
	switch {
	case gb < 20:
		c.report(statusFail, "Disk space", msg+", minimum is 20GB")
	case gb < 50:
		c.report(statusWarn, "Disk space", msg+", recommended is 50GB+")
	default:
		c.report(statusOK, "Disk space", msg)
	}
}

func checkSwap(c *checker) {
	mem := readMeminfo()
	if swap := mem["SwapTotal"]; swap > 0 {
		c.report(statusWarn, "Swap", fmt.Sprintf("swap is enabled (%.1fGB), disabling it is recommended (swapoff -a)", float64(swap)/(1<<30)))
	} else {
		c.report(statusOK, "Swap", "disabled")
	}
}

var hostnameRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

func checkHostname(c *checker) {
	h, err := os.Hostname()
	if err != nil {
		c.report(statusFail, "Hostname", err.Error())
		return
	}
	if len(h) > 63 || !hostnameRe.MatchString(h) {
		c.report(statusFail, "Hostname", fmt.Sprintf("%q is not a valid RFC 1123 name (lowercase, digits, '-', '.')", h))
		return
	}
	if h == "localhost" {
		c.report(statusFail, "Hostname", "hostname must not be localhost")
		return
	}
	if _, err := net.LookupHost(h); err != nil {
		c.report(statusWarn, "Hostname", fmt.Sprintf("%s does not resolve (add it to /etc/hosts)", h))
		return
	}
	c.report(statusOK, "Hostname", h+" (must be unique in the cluster)")
}

func checkCgroups(c *checker) {
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
		c.report(statusOK, "Cgroups", "v2")
		return
	}
	if _, err := os.Stat("/sys/fs/cgroup"); err == nil {
		c.report(statusOK, "Cgroups", "v1")
		return
	}
	c.report(statusFail, "Cgroups", "/sys/fs/cgroup not found")
}

func moduleAvailable(name string) (loaded bool, available bool) {
	if _, err := os.Stat("/sys/module/" + name); err == nil {
		return true, true
	}
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false, false
	}
	release := strings.TrimSpace(string(data))
	for _, f := range []string{"modules.dep", "modules.builtin"} {
		data, err := os.ReadFile("/lib/modules/" + release + "/" + f)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "/"+name+".ko") {
			return false, true
		}
	}
	return false, false
}

func checkKernelModules(c *checker) {
	for _, m := range []string{"overlay", "br_netfilter"} {
		loaded, available := moduleAvailable(m)
		switch {
		case loaded:
			c.report(statusOK, "Kernel module "+m, "loaded")
		case available:
			c.report(statusOK, "Kernel module "+m, "available (RKE2 loads it on start)")
		default:
			c.report(statusWarn, "Kernel module "+m, "not found, check your kernel")
		}
	}
}

func checkIPForward(c *checker) {
	data, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err != nil {
		c.report(statusWarn, "IP forwarding", err.Error())
		return
	}
	if strings.TrimSpace(string(data)) == "1" {
		c.report(statusOK, "IP forwarding", "enabled")
	} else {
		c.report(statusInfo, "IP forwarding", "disabled (RKE2 enables it on start)")
	}
}

func checkSELinux(c *checker) {
	data, err := os.ReadFile("/sys/fs/selinux/enforce")
	if err != nil {
		c.report(statusOK, "SELinux", "not enabled")
		return
	}
	if strings.TrimSpace(string(data)) != "1" {
		c.report(statusOK, "SELinux", "permissive")
		return
	}
	if fileExists("/usr/share/selinux/packages/rke2.pp") || commandSucceeds("rpm", "-q", "rke2-selinux") {
		c.report(statusOK, "SELinux", "enforcing, rke2-selinux installed")
	} else {
		c.report(statusWarn, "SELinux", "enforcing, rke2-selinux will be installed by install script (needs container-selinux)")
	}
}

func checkExistingInstallations(c *checker) {
	found := []string{}
	for _, p := range []string{
		"/etc/rancher/rke2/config.yaml",
		"/var/lib/rancher/rke2/server/db",
		"/var/lib/rancher/rke2/agent",
		"/usr/local/bin/rke2",
		"/usr/bin/rke2",
	} {
		if fileExists(p) {
			found = append(found, p)
		}
	}
	if len(found) > 0 {
		c.report(statusWarn, "Existing RKE2", "found "+strings.Join(found, ", "))
	} else {
		c.report(statusOK, "Existing RKE2", "not found")
	}

	if fileExists("/usr/local/bin/k3s") || fileExists("/var/lib/rancher/k3s") {
		c.report(statusFail, "Existing K3s", "K3s found, uninstall it first (k3s-uninstall.sh)")
	} else {
		c.report(statusOK, "Existing K3s", "not found")
	}
}

func checkConflictingServices(c *checker) {
	if !hasSystemd() {
		c.report(statusWarn, "Systemd", "systemctl not found, RKE2 requires systemd")
		return
	}
	c.report(statusOK, "Systemd", "available")

	conflicts := map[string]string{
		"docker":         "Docker is running, may conflict with RKE2 containerd/iptables",
		"containerd":     "standalone containerd is running, may conflict with RKE2 containerd",
		"firewalld":      "firewalld is active, RKE2 recommends disabling it",
		"ufw":            "ufw is active, open required ports or disable it",
		"kubelet":        "kubelet service is running, another Kubernetes is installed",
		"k3s":            "K3s is running",
		"nm-cloud-setup": "nm-cloud-setup is enabled, it must be disabled (breaks CNI routing)",
	}
	for _, svc := range []string{"docker", "containerd", "kubelet", "k3s", "firewalld", "ufw", "nm-cloud-setup"} {
		active := commandSucceeds("systemctl", "is-active", "--quiet", svc)
		enabled := svc == "nm-cloud-setup" && (commandSucceeds("systemctl", "is-enabled", "--quiet", "nm-cloud-setup.service") ||
			commandSucceeds("systemctl", "is-enabled", "--quiet", "nm-cloud-setup.timer"))
		if !active && !enabled {
			continue
		}
		s := statusWarn
		if svc == "kubelet" || svc == "k3s" || svc == "nm-cloud-setup" {
			s = statusFail
		}
		c.report(s, "Service "+svc, conflicts[svc])
	}
}

func checkNetworkManager(c *checker) {
	if !commandSucceeds("systemctl", "is-active", "--quiet", "NetworkManager") {
		return
	}
	files, _ := os.ReadDir("/etc/NetworkManager/conf.d")
	for _, f := range files {
		data, err := os.ReadFile("/etc/NetworkManager/conf.d/" + f.Name())
		if err == nil && strings.Contains(string(data), "unmanaged-devices") &&
			(strings.Contains(string(data), "cali") || strings.Contains(string(data), "flannel")) {
			c.report(statusOK, "NetworkManager", "CNI interfaces are unmanaged ("+f.Name()+")")
			return
		}
	}
	c.report(statusWarn, "NetworkManager",
		"active, create /etc/NetworkManager/conf.d/rke2-canal.conf with: "+
			"[keyfile] unmanaged-devices=interface-name:flannel*;interface-name:cali*;interface-name:tunl*;interface-name:vxlan.calico;interface-name:vxlan-v6.calico;interface-name:wireguard.cali;interface-name:wg-v6.cali")
}

func checkTimeSync(c *checker) {
	if _, err := exec.LookPath("timedatectl"); err != nil {
		c.report(statusWarn, "Time sync", "timedatectl not found, cannot verify")
		return
	}
	out, err := exec.Command("timedatectl", "show", "-p", "NTPSynchronized", "--value").Output()
	if err != nil {
		c.report(statusWarn, "Time sync", "cannot verify: "+err.Error())
		return
	}
	if strings.TrimSpace(string(out)) == "yes" {
		c.report(statusOK, "Time sync", "NTP synchronized")
	} else {
		c.report(statusWarn, "Time sync", "clock is not NTP synchronized (important for etcd and certificates)")
	}
}

func checkBinaries(c *checker) {
	missing := []string{}
	for _, b := range []string{"tar", "sha256sum"} {
		if _, err := exec.LookPath(b); err != nil {
			missing = append(missing, b)
		}
	}
	_, errCurl := exec.LookPath("curl")
	_, errWget := exec.LookPath("wget")
	if errCurl != nil && errWget != nil {
		missing = append(missing, "curl or wget")
	}
	if len(missing) > 0 {
		c.report(statusFail, "Required tools", "missing: "+strings.Join(missing, ", "))
	} else {
		c.report(statusOK, "Required tools", "tar, sha256sum, curl/wget")
	}

	if _, err := exec.LookPath("iptables"); err != nil {
		c.report(statusInfo, "iptables", "not found on host (RKE2 ships its own)")
	}
}

func checkDefaultRoute(c *checker) {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		c.report(statusWarn, "Default route", err.Error())
		return
	}
	for _, line := range strings.Split(string(data), "\n")[1:] {
		f := strings.Fields(line)
		if len(f) > 2 && f[1] == "00000000" {
			c.report(statusOK, "Default route", "via "+f[0])
			return
		}
	}
	c.report(statusFail, "Default route", "no default route, RKE2 requires one (add a dummy one for airgap)")
}

func checkPorts(c *checker, role string) {
	tcp := []int{10250}
	if role == "server" {
		tcp = append(tcp, 6443, 9345, 2379, 2380, 2381, 10257, 10259)
	}
	busy := []string{}
	for _, p := range tcp {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err != nil {
			busy = append(busy, strconv.Itoa(p))
			continue
		}
		l.Close()
	}
	for _, p := range []int{8472} {
		l, err := net.ListenPacket("udp", fmt.Sprintf(":%d", p))
		if err != nil {
			busy = append(busy, strconv.Itoa(p)+"/udp")
			continue
		}
		l.Close()
	}
	if len(busy) > 0 {
		c.report(statusFail, "Ports", "already in use: "+strings.Join(busy, ", "))
	} else {
		c.report(statusOK, "Ports", "required ports are free")
	}
}

func checkInternet(c *checker) {
	client := &http.Client{Timeout: 10 * time.Second}
	for _, u := range []string{
		"https://get.rke2.io",
		"https://github.com",
		"https://registry-1.docker.io/v2/",
	} {
		resp, err := client.Head(u)
		if err != nil {
			c.report(statusFail, "Connectivity", u+": "+err.Error())
			continue
		}
		resp.Body.Close()
		c.report(statusOK, "Connectivity", fmt.Sprintf("%s (%d)", u, resp.StatusCode))
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func hasSystemd() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil && fileExists("/run/systemd/system")
}

func commandSucceeds(name string, args ...string) bool {
	if _, err := exec.LookPath(name); err != nil {
		return false
	}
	return exec.Command(name, args...).Run() == nil
}
