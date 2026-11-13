package check_cve

import (
	"fmt"
	"os"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "check-cve",
	Short: "Check if this server is affected by a CVE (exit 0 not affected, 1 vulnerable/unknown)",
}

type Verdict int

const (
	NotAffected Verdict = iota
	Mitigated
	Vulnerable
	Unknown
)

func (v Verdict) String() string {
	switch v {
	case NotAffected:
		return "NOT AFFECTED"
	case Mitigated:
		return "MITIGATED"
	case Vulnerable:
		return "VULNERABLE"
	default:
		return "UNKNOWN"
	}
}

type LogFunc func(format string, a ...any)

// NewCveCmd creates a check-cve subcommand for one CVE.
func NewCveCmd(id, description string, check func(log LogFunc) (Verdict, string)) *cobra.Command {
	return &cobra.Command{
		Use:   strings.ToLower(id),
		Short: description,
		Args:  cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			fmt.Printf("Checking %s: %s\n\n", strings.ToUpper(id), description)
			v, reason := check(func(format string, a ...any) {
				fmt.Printf("  - "+format+"\n", a...)
			})
			fmt.Printf("\nResult: %s (%s)\n", v, reason)
			if v == NotAffected || v == Mitigated {
				os.Exit(0)
			}
			os.Exit(1)
		},
	}
}
