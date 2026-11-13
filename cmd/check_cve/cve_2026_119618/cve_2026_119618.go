package cve_2026_119618

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sikalabs/slr/cmd/check_cve"
)

// CVE-2026-119618: Apache HTTP Server mod_http2 race condition RCE.
// Affected: httpd 2.4.55 - 2.4.70 with mod_http2 loaded and h2/h2c enabled
// in the Protocols directive. Fixed upstream in 2.4.71. Distributions backport
// the fix without changing the version, so package changelogs are checked too.

const cveID = "CVE-2026-119618"

func init() {
	check_cve.Cmd.AddCommand(check_cve.NewCveCmd(
		cveID,
		"Apache HTTP Server mod_http2 race condition RCE (2.4.55 - 2.4.70)",
		check,
	))
}

var apacheVersionRe = regexp.MustCompile(`Apache/(\d+)\.(\d+)\.(\d+)`)

func check(log check_cve.LogFunc) (check_cve.Verdict, string) {
	bin, ctl := findApache()
	if bin == "" {
		log("Apache httpd binary not found (httpd, apache2)")
		return check_cve.NotAffected, "Apache httpd is not installed on this host (containers are not checked)"
	}
	log("Apache binary: %s", bin)

	out, err := exec.Command(bin, "-v").CombinedOutput()
	m := apacheVersionRe.FindStringSubmatch(string(out))
	if err != nil && m == nil {
		log("cannot get version: %s", err)
		return check_cve.Unknown, "cannot determine Apache version"
	}
	if m == nil {
		log("cannot parse version from: %s", strings.TrimSpace(string(out)))
		return check_cve.Unknown, "cannot determine Apache version"
	}
	ver := [3]int{atoi(m[1]), atoi(m[2]), atoi(m[3])}
	log("Apache version: %d.%d.%d", ver[0], ver[1], ver[2])

	if cmpVersion(ver, [3]int{2, 4, 55}) < 0 || cmpVersion(ver, [3]int{2, 4, 71}) >= 0 {
		return check_cve.NotAffected, "version is outside of affected range 2.4.55 - 2.4.70"
	}
	log("version is in affected range 2.4.55 - 2.4.70")

	if pkg, patched := packageHasFix(cveID); pkg != "" {
		log("package: %s", pkg)
		if patched {
			log("package changelog mentions %s", cveID)
			return check_cve.NotAffected, "fix is backported by the distribution package"
		}
		log("package changelog does not mention %s", cveID)
	}

	loaded, known := http2ModuleLoaded(ctl, bin)
	if !known {
		log("cannot list loaded modules (run as root?)")
		return check_cve.Unknown, "affected version, cannot verify if mod_http2 is loaded, upgrade to 2.4.71+"
	}
	if !loaded {
		log("mod_http2 is not loaded")
		return check_cve.Mitigated, "affected version, but mod_http2 is not loaded, upgrade anyway"
	}
	log("mod_http2 is loaded")

	files := h2ProtocolFiles()
	if len(files) == 0 {
		log("no Protocols directive with h2/h2c found in config")
		return check_cve.Mitigated, "affected version, but HTTP/2 is not enabled in Protocols, upgrade anyway"
	}
	for _, f := range files {
		log("HTTP/2 enabled in %s", f)
	}
	return check_cve.Vulnerable, "upgrade Apache to 2.4.71+ or the patched distro package, or set Protocols http/1.1 as a workaround"
}

func findApache() (bin, ctl string) {
	for _, b := range []string{"httpd", "apache2"} {
		if p, err := exec.LookPath(b); err == nil {
			bin = p
			break
		}
	}
	if bin == "" {
		for _, p := range []string{"/usr/sbin/httpd", "/usr/sbin/apache2", "/usr/local/apache2/bin/httpd"} {
			if _, err := os.Stat(p); err == nil {
				bin = p
				break
			}
		}
	}
	for _, c := range []string{"apache2ctl", "apachectl", "/usr/sbin/apache2ctl", "/usr/sbin/apachectl", "/usr/local/apache2/bin/apachectl"} {
		if p, err := exec.LookPath(c); err == nil {
			ctl = p
			break
		}
	}
	return bin, ctl
}

// packageHasFix returns the installed package name/version and whether its
// changelog mentions the CVE (= the fix is backported).
func packageHasFix(cve string) (pkg string, patched bool) {
	if _, err := exec.LookPath("rpm"); err == nil {
		if out, err := exec.Command("rpm", "-q", "httpd").Output(); err == nil {
			pkg = strings.TrimSpace(string(out))
			cl, _ := exec.Command("rpm", "-q", "--changelog", "httpd").Output()
			return pkg, strings.Contains(string(cl), cve)
		}
	}
	if _, err := exec.LookPath("dpkg-query"); err == nil {
		if out, err := exec.Command("dpkg-query", "-W", "-f", "${Package} ${Version}", "apache2").Output(); err == nil {
			pkg = strings.TrimSpace(string(out))
			for _, name := range []string{"apache2", "apache2-bin"} {
				if fileContainsGz("/usr/share/doc/"+name+"/changelog.Debian.gz", cve) {
					return pkg, true
				}
			}
			return pkg, false
		}
	}
	return "", false
}

func fileContainsGz(path, needle string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	data, err := io.ReadAll(gz)
	return err == nil && strings.Contains(string(data), needle)
}

func http2ModuleLoaded(ctl, bin string) (loaded, known bool) {
	cmds := [][]string{}
	if ctl != "" {
		cmds = append(cmds, []string{ctl, "-M"})
	}
	cmds = append(cmds, []string{bin, "-M"})
	for _, c := range cmds {
		out, _ := exec.Command(c[0], c[1:]...).CombinedOutput()
		s := string(out)
		if !strings.Contains(s, "Loaded Modules") {
			continue
		}
		return strings.Contains(s, "http2_module"), true
	}
	return false, false
}

// h2ProtocolFiles returns config files containing an active Protocols
// directive with h2 or h2c. Directories *-available (Debian) are skipped.
func h2ProtocolFiles() []string {
	found := []string{}
	for _, dir := range []string{"/etc/httpd", "/etc/apache2", "/usr/local/apache2/conf"} {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if strings.HasSuffix(d.Name(), "-available") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".conf") && !strings.HasSuffix(path, ".load") {
				return nil
			}
			if fileEnablesH2(path) {
				found = append(found, path)
			}
			return nil
		})
	}
	return found
}

func fileEnablesH2(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) < 2 || !strings.EqualFold(fields[0], "Protocols") {
			continue
		}
		for _, p := range fields[1:] {
			if p == "h2" || p == "h2c" {
				return true
			}
		}
	}
	return false
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func cmpVersion(a, b [3]int) int {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}
