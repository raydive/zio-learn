package playground

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func scalaCommand(ctx context.Context, name, dir string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "scala-cli", "run", name, "--server=false", "--jvm", "system")
	cmd.Dir = dir
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Kill the entire process group, including compiler and application JVMs.
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

func removeDirectory(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		slog.Warn("remove temporary directory", "path", dir, "error", err)
	}
}
