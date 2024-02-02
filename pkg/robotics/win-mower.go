package robotics

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
)

type WinMowerRegistry struct {
	cacheDir string
}

type WinMower struct {
	Path string
}

func NewWinMowerRegistry(cacheDir string) *WinMowerRegistry {
	return &WinMowerRegistry{
		cacheDir: cacheDir,
	}
}

func (w *WinMowerRegistry) GetCachedWinMower(platform Platform, ctx context.Context) (*WinMower, error) {
	var wmDir string
	err := filepath.WalkDir(w.cacheDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && d.Name() == platform.String() {
			wmDir = path
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if wmDir == "" {
		return nil, fmt.Errorf("no cached WinMower found for platform %s", platform)
	}

	path, err := locateWinMowerExecutable(wmDir)
	if err != nil {
		return nil, err
	}

	return &WinMower{
		Path: path,
	}, nil
}

func locateWinMowerExecutable(dir string) (string, error) {
	var exePath string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if filepath.Ext(path) == ".exe" {
			exePath = path
			return nil
		}
		return nil
	})
	if exePath == "" {
		return "", fmt.Errorf("no exe found in %s", dir)
	}
	return exePath, err
}
