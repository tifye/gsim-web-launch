package robotics

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/Tifufu/gsim-web-launch/pkg/ext"
	"github.com/charmbracelet/log"
)

type WinMowerRegistry struct {
	CacheDir       string
	bundleRegistry *BundleRegistry
}

type WinMower struct {
	Path string
}

func NewWinMowerRegistry(cacheDir string, bregsitry *BundleRegistry) *WinMowerRegistry {
	return &WinMowerRegistry{
		bundleRegistry: bregsitry,
		CacheDir:       cacheDir,
	}
}

func (w *WinMowerRegistry) GetWinMower(platform Platform, ctx context.Context) (*WinMower, error) {
	wm, err := w.GetCachedWinMower(platform, ctx)
	if err != nil {
		return nil, err
	}
	if wm != nil {
		log.Debug("Using cached winmower")
		return wm, nil
	}

	btypes, err := w.bundleRegistry.FetchBundleTypes(ctx)
	if err != nil {
		return nil, err
	}
	log.Debugf("Found %d bundle types\n", len(btypes))

	btypes = FilterBundleTypes(btypes, platform)
	if len(btypes) == 0 {
		return nil, fmt.Errorf("no bundle types found for platform %s", platform)
	}
	log.Debugf("Found %d bundle types for platform %s\n", len(btypes), platform)

	// Endpoint returns them sorted by date (i think)
	latestType := btypes[0]
	log.Debugf("Latest bundle type: %s\n", latestType.Name)

	latestBuild, err := w.bundleRegistry.FetchLatestRelease(ctx, latestType.Name)
	if err != nil {
		return nil, err
	}
	log.Debugf("Latest build: %s\n", latestBuild.BlobUrl)

	dir := filepath.Join(w.CacheDir, platform.String())
	req, err := http.NewRequestWithContext(ctx, "GET", latestBuild.BlobUrl, nil)
	if err != nil {
		return nil, err
	}
	AddTifAuthHeaders(req)
	log.Debug("Downloading and unpacking winmower...")
	err = ext.DownloadAndUnpack(req, dir)
	if err != nil {
		return nil, err
	}

	wmPath, err := locateWinMowerExecutable(dir)
	if err != nil {
		return nil, err
	}

	return &WinMower{
		Path: wmPath,
	}, nil
}

func (w *WinMowerRegistry) GetCachedWinMower(platform Platform, ctx context.Context) (*WinMower, error) {
	var wmDir string
	err := filepath.WalkDir(w.CacheDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && d.Name() == platform.String() {
			wmDir = path
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if wmDir == "" {
		return nil, nil
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
