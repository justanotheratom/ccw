package cmd

import (
	"github.com/ccw/ccw/internal/workspace"
)

func newManager() (*workspace.Manager, error) {
	return workspace.NewManager("", nil)
}
