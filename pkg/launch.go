package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
)

func RunTestBundle(bundlePath string) {
	execPath := `C:\Users\demat\AppData\Local\TifApp\TifConsole.Auto.exe`
	args := []string{
		// `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GSP_190703524\P25_190703524.zip`,
		bundlePath,
		"-tcpAddress",
		"127.0.0.1:4250",
		"-output",
		`D:\Projects\_work\_pocs\gsim-web-launch\testdata`,
	}
	cmd := exec.Command(execPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Error("Failed to launch test bundle", "err", err)
		panic(err)
	}
}

func LaunchSimulator(mapPath string) {
	exePath := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GardenSimulator\GardenSimulator.exe`
	args := []string{
		// "-config", `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GSP_190703524\map.json`,
		"-config", mapPath,
		"-log", "true",
		"-time-scale", "1",
		"-screen-width", "1280",
		"-screen-height", "720",
		"-quality-level", "6",
	}
	cmd := exec.Command(exePath, args...)
	err := cmd.Start()
	if err != nil {
		log.Error("Failed to launch simulator", "err", err)
		panic(err)
	}
}

func LaunchWinMower(exePath, platform string) {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		log.Error("Failed to get user cache dir", "err", err)
		panic(err)
	}
	wmDir := filepath.Join(cachedir, "gsim/winmower-filesystems", platform)
	err = os.MkdirAll(wmDir, 0755)
	if err != nil {
		log.Error("Failed to create winmower dir", "err", err)
		panic(err)
	}

	cmd := exec.Command(exePath)
	cmd.Dir = wmDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}

	test := &Test{}
	cmd.Stdout = test
	cmd.Stderr = test
	err = cmd.Start()
	if err != nil {
		log.Error("Failed to launch winmower", "err", err)
		panic(err)
	}
}

type Test struct {
}

func (t *Test) Write(bytes []byte) (int, error) {
	str := string(bytes)
	str = strings.TrimSuffix(str, "\n")
	switch {
	case strings.Contains(str, "ERROR"):
		log.Error(str)
	case strings.Contains(str, "WARNING"):
		log.Warn(str)
	case strings.Contains(str, "INFO"):
		log.Info(str)
	case strings.Contains(str, "DEBUG"):
		log.Debug(str)
	default:
		log.Info(str)
	}
	return len(bytes), nil
}
