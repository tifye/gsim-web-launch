package runner

import (
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
)

type TestBundleRunner struct {
	tifConsolePath string
	logger         io.Writer
}

type TifConsoleLogger struct {
	logger *log.Logger
}

func NewTestBundleRunner(tifConsolePath string, logger io.Writer) *TestBundleRunner {
	return &TestBundleRunner{
		tifConsolePath: tifConsolePath,
		logger:         logger,
	}
}

func NewTifConsoleLogger(logger *log.Logger) *TifConsoleLogger {
	return &TifConsoleLogger{
		logger: logger,
	}
}

func (r *TestBundleRunner) Run(ctx context.Context, bundlePath string, args ...string) error {
	cmdArgs := append([]string{bundlePath}, args...)
	cmd := exec.CommandContext(ctx, r.tifConsolePath, cmdArgs...)
	cmd.Stdout = r.logger
	cmd.Stderr = r.logger
	return cmd.Run()
}

func (l *TifConsoleLogger) Write(bytes []byte) (int, error) {
	str := string(bytes)
	str = strings.TrimSuffix(str, "\n")
	l.logger.Print(str)
	return len(bytes), nil
}
