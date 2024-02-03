package runner

import (
	"context"
	"os/exec"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
)

type WinMowerRunner struct {
	dir    string
	path   string
	logger *WinMowerLogger
	cmd    *exec.Cmd
}

type WinMowerLogger struct {
	logger *log.Logger
}

func NewWinMowerRunner(dir, path string, logger *WinMowerLogger) *WinMowerRunner {
	return &WinMowerRunner{
		dir:    dir,
		path:   path,
		logger: logger,
	}
}

func (r *WinMowerRunner) Start(ctx context.Context) error {
	r.cmd = exec.CommandContext(ctx, r.path)
	r.cmd.Dir = r.dir
	r.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	r.cmd.Stdout = r.logger
	r.cmd.Stderr = r.logger
	return r.cmd.Start()
}

func (r *WinMowerRunner) Stop() error {
	return r.cmd.Cancel()
}

func NewWinMowerLogger(logger *log.Logger) *WinMowerLogger {
	return &WinMowerLogger{
		logger: logger,
	}
}

func (r *WinMowerLogger) Write(bytes []byte) (int, error) {
	str := string(bytes)
	str = strings.TrimSuffix(str, "\n")

	lines := strings.Split(str, "\n")
	first := lines[0]

	switch {
	case strings.Contains(first, "ERROR"):
		r.logger.Error(first)
	case strings.Contains(first, "WARNING"):
		r.logger.Warn(first)
	case strings.Contains(first, "INFO"):
		r.logger.Info(first)
	case strings.Contains(first, "DEBUG"):
		r.logger.Debug(first)
	default:
		r.logger.Info(first)
	}

	rest := lines[1:]
	for _, line := range rest {
		r.Write([]byte(line))
	}

	return len(bytes), nil
}
