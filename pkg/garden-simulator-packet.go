package pkg

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type GSPPaths struct {
	Map        string
	TestBundle string
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
