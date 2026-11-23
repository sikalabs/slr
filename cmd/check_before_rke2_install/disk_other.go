//go:build !linux

package check_before_rke2_install

import (
	"fmt"
	"runtime"
)

func diskAvailable(path string) (uint64, error) {
	return 0, fmt.Errorf("not supported on %s", runtime.GOOS)
}
