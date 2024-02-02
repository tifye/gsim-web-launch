package ext

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadAndUnpack(req *http.Request, dest string) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("response failed with %s", resp.Status)
		}
		return fmt.Errorf("response failed with %s, %s", resp.Status, string(b))
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "winmower_*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}

	err = Unzip(tmpFile.Name(), dest)
	if err != nil {
		return err
	}

	return nil
}

func Unzip(zipFile string, dest string) error {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		outputPath := filepath.Join(dest, file.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(outputPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", outputPath)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(outputPath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			return err
		}

		outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer outputFile.Close()

		archiveFile, err := file.Open()
		if err != nil {
			return err
		}
		defer archiveFile.Close()

		if _, err := io.Copy(outputFile, archiveFile); err != nil {
			return err
		}
	}

	return nil
}
