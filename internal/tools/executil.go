package tools

import (
	"context"
	"os/exec"
)

// ExecCommandContext is a small wrapper for testability.
func ExecCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
