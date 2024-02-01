package pkg

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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
		log.Fatalf("Failed to launch test bundle: %s", err)
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
		log.Fatalf("Failed to launch simulator: %s", err)
	}
}

func LaunchWinMower(exePath, platform string) {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("Failed to get user cache dir: %s", err)
	}
	wmDir := filepath.Join(cachedir, "winmower", platform)
	err = os.MkdirAll(wmDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create winmower dir: %s", err)
	}

	//exePath := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\40.x_Main-App-P25-Win_master_build-240131_153656\40.x_Main-App-P25-Win_master_build-240131_153656.exe`
	cmd := exec.Command("cmd.exe", "/C", "start", exePath)
	cmd.Dir = wmDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to launch winmower: %s", err)
	}
}
