package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd/registry"
	"github.com/Tifufu/gsim-web-launch/pkg"
	"github.com/spf13/cobra"
)

var (
	serialNumber string
	platform     string
	testingDir   string = filepath.Join(`D:\Projects\_work\_pocs\gsim-web-launch\_vendor`)
)

var rootCmd = &cobra.Command{
	Use:   "gsim-web-launch",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if r := recover(); r != nil {
				log.Fatalf("Failed to launch: %s", r)
				time.Sleep(5 * time.Second)
			}
			time.Sleep(5 * time.Second)
		}()

		_, err := downloadAndUnpackWinMower(platform)
		if err != nil {
			log.Fatalf("Failed to download and unpack WinMower: %s", err)
		}

		gspPaths, err := downloadAndUnpackGSP(serialNumber, platform)
		if err != nil {
			log.Fatalf("Failed to download and unpack GSP: %s", err)
		}

		log.Printf("\nMap: %s\nTestBundle: %s\n", gspPaths.Map, gspPaths.TestBundle)

		// fmt.Println("Launching winmower...")
		// pkg.LaunchWinMower(winMowerPath)
		// fmt.Println("Launching simulator...")
		// pkg.LaunchSimulator(gspPaths.Map)
		// time.Sleep(5 * time.Second)
		// fmt.Println("Launching test bundle...")
		// pkg.RunTestBundle(gspPaths.TestBundle)
	},
}

func init() {
	cobra.MousetrapHelpText = ""
	rootCmd.AddCommand(registry.RegistryCmd)

	rootCmd.Flags().StringVarP(&serialNumber, "serial-number", "s", "", "Serial number of the device")
	rootCmd.MarkFlagRequired("serial-number")

	rootCmd.Flags().StringVarP(&platform, "platform", "p", "P25", "Platform of the device")
	rootCmd.MarkFlagRequired("platform")
}

func Execute(args []string) {
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func downloadAndUnpackWinMower(platform string) (string, error) {
	types, err := pkg.FetchBundleTypes()
	if err != nil {
		return "", err
	}

	plat := pkg.Platform(platform)
	plat.Set(platform)
	types = pkg.FilterBundleTypes(types, plat)
	if len(types) == 0 {
		return "", fmt.Errorf("no bundle types found for platform %s", platform)
	}
	latestType := types[0]

	latestBuild, err := pkg.FetchLatestRelease(latestType.Name)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(testingDir, "winmower", latestType.Name)
	err = pkg.DownloadAndUnpack(latestBuild.BlobUrl, dir)
	if err != nil {
		return "", err
	}

	var exePath string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".exe" {
			exePath = path
			return nil
		}
		return nil
	})
	if exePath == "" {
		return "", fmt.Errorf("no exe found in %s", dir)
	}
	log.Printf("Found WinMower exe at %s", exePath)

	return exePath, err
}

func downloadAndUnpackGSP(serialNumber, platform string) (*pkg.GSPPaths, error) {
	path := filepath.Join(testingDir, "gsp", "GSP_"+serialNumber)
	baseUrl := os.Getenv("GSP_API")
	endpoint := fmt.Sprintf("%s/packet/%s/%s", baseUrl, serialNumber, platform)
	err := pkg.DownloadAndUnpack(endpoint, path)
	if err != nil {
		return nil, err
	}

	gsp, err := pkg.LocateGSPPaths(path, serialNumber)
	if err != nil {
		return nil, err
	}

	return gsp, nil
}
