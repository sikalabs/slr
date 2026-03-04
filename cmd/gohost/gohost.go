package gohost

import (
	gohostcmd "github.com/sikalabsx/gohost/cmd/root"
	"github.com/sikalabs/slr/cmd/root"
)

func init() {
	root.Cmd.AddCommand(gohostcmd.Cmd)
}
