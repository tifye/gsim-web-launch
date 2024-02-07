package robotics

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tifufu/gsim-web-launch/pkg/ext"
	"github.com/charmbracelet/log"
)

type GSPRegistry struct {
	cacheDir string
	baseUrl  string
}

type GSPPaths struct {
	Map        string
	TestBundle string
}

func NewGSPRegistry(cacheDir, baseUrl string) *GSPRegistry {
	return &GSPRegistry{
		cacheDir: cacheDir,
		baseUrl:  baseUrl,
	}
}

func (r *GSPRegistry) GetGSP(serialNumber, platform string) (*GSPPaths, error) {
	gsp, err := r.GetGSPFromCache(serialNumber)
	if err != nil {
		return nil, err
	}
	if gsp != nil {
		log.Debug("Using cached GSP")
		return gsp, nil
	}

	endpoint := fmt.Sprintf("%s/packet/%s/%s", r.baseUrl, serialNumber, platform)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	AddTifAuthHeaders(req)

	dir := filepath.Join(r.cacheDir, serialNumber)
	err = ext.DownloadAndUnpack(req, dir)
	if err != nil {
		return nil, err
	}

	gsp, err = LocateGSPPaths(dir, serialNumber)
	if err != nil {
		return nil, err
	}

	return gsp, nil
}

func (r *GSPRegistry) GetGSPFromCache(serialNumber string) (*GSPPaths, error) {
	dir := filepath.Join(r.cacheDir, serialNumber)
	_, err := os.Stat(dir)
	if err == nil {
		return LocateGSPPaths(dir, serialNumber)
	}
	if os.IsNotExist(err) {
		return nil, nil
	}
	return nil, err
}

func LocateGSPPaths(dir string, serialNumber string) (*GSPPaths, error) {
	gspPaths := &GSPPaths{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		switch {
		case info.Name() == "map.json":
			gspPaths.Map = path
		case strings.HasSuffix(info.Name(), serialNumber+".zip"):
			gspPaths.TestBundle = path
		default:
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if gspPaths.Map == "" || gspPaths.TestBundle == "" {
		return nil, errors.New("Failed to locate GSP paths")
	}

	return gspPaths, nil
}
