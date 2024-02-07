package robotics

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/Tifufu/gsim-web-launch/pkg/ext"
	"github.com/charmbracelet/log"
)

type SimulatorRegistry struct {
	cacheDir       string
	bundleRegistry *BundleRegistry
}

type Simulator struct {
	Path string
}

func NewSimulatorRegistry(cacheDir string, bregsitry *BundleRegistry) *SimulatorRegistry {
	return &SimulatorRegistry{
		bundleRegistry: bregsitry,
		cacheDir:       cacheDir,
	}
}

func (s *SimulatorRegistry) GetSimulator(ctx context.Context) (*Simulator, error) {
	sim, err := s.GetCachedSimulator(ctx)
	if err != nil {
		return nil, err
	}
	if sim != nil {
		log.Info("Using cached simulator")
		return sim, nil
	}

	log.Info("Fetching simulator...")
	latestBuild, err := s.bundleRegistry.FetchLatestRelease(ctx, "GardenSimulator")
	if err != nil {
		return nil, err
	}
	log.Printf("Latest Simulator build: %s\n", latestBuild.BlobUrl)

	req, err := http.NewRequestWithContext(ctx, "GET", latestBuild.BlobUrl, nil)
	if err != nil {
		return nil, err
	}
	AddTifAuthHeaders(req)
	log.Info("Downloading and unpacking simulator...")
	err = ext.DownloadAndUnpack(req, s.cacheDir)
	if err != nil {
		return nil, err
	}

	return s.GetCachedSimulator(ctx)
}

func (s *SimulatorRegistry) GetCachedSimulator(ctx context.Context) (*Simulator, error) {
	var exePath string
	err := filepath.Walk(s.cacheDir, func(path string, info fs.FileInfo, err error) error {
		if filepath.Base(path) == "GardenSimulator.exe" {
			exePath = path
			return nil
		}
		return nil
	})

	if errors.Is(err, fs.ErrNotExist) || exePath == "" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &Simulator{
		Path: exePath,
	}, nil
}
