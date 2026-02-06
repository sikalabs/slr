package example

import (
	_ "github.com/sikalabs/scr/cmd"
	scr_root "github.com/sikalabs/scr/cmd/root"
	"github.com/sikalabs/slr/cmd/root"
)

func init() {
	root.Cmd.AddCommand(scr_root.Cmd)
}
