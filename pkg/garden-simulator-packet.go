package pkg

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type GSPPaths struct {
	Map        string
	TestBundle string
}

func DownloadGSP(serialNumber, model, dest string) error {
	apiKey := os.Getenv("API_KEY")
	token := os.Getenv("TOKEN")
	baseUrl := os.Getenv("GSP_API")

	url := fmt.Sprintf("%s/packet/%s/%s", baseUrl, serialNumber, model)
	log.Printf("Downloading GSP from %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", "fuit-pie")
	req.Header.Set("token", token)
	req.Header.Set("Ocp-Apim-Subscription-Key", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("Failed to download GSP: %s", res.Status)
		return errors.New("Failed to download GSP")
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}

	return nil
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
