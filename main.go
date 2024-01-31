package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Launching winmower...")
	launchWinMower()
	fmt.Println("Launching simulator...")
	launchSimulator()
	time.Sleep(3 * time.Second)
	fmt.Println("Launching test bundle...")
	runTestBundle()
}

func runTestBundle() {
	execPath := `C:\Users\demat\AppData\Local\TifApp\TifConsole.Auto.exe`
	bundleZipPath := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GSP_190703524\P25_190703524.zip`
	args := []string{
		bundleZipPath,
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

func launchSimulator() {
	exePath := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GardenSimulator\GardenSimulator.exe`
	args := []string{
		"-config", `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GSP_190703524\map.json`,
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

func launchWinMower() {
	exePath := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\40.x_Main-App-P25-Win_master_build-240131_153656\40.x_Main-App-P25-Win_master_build-240131_153656.exe`
	cmd := exec.Command("cmd.exe", "/C", "start", exePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to launch winmower: %s", err)
	}
}
