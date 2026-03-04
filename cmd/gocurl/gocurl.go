package gocurl

import (
	gocurlcmd "github.com/sikalabsx/gocurl/cmd/root"
	"github.com/sikalabs/slr/cmd/root"
)

func init() {
	root.Cmd.AddCommand(gocurlcmd.Cmd)
}
