package runner

import "os/exec"

func LaunchSimulator(simPath string, mapPath string) error {
	args := []string{
		"-config", mapPath,
		"-log", "true",
		"-time-scale", "1",
		"-screen-width", "1280",
		"-screen-height", "720",
		"-quality-level", "6",
	}
	cmd := exec.Command(simPath, args...)
	return cmd.Start()
}
