package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd/cli"
	"github.com/Tifufu/gsim-web-launch/cmd/registry"
	"github.com/Tifufu/gsim-web-launch/pkg"
	"github.com/Tifufu/gsim-web-launch/pkg/ext"
	"github.com/Tifufu/gsim-web-launch/pkg/robotics"
	"github.com/spf13/cobra"
)

var (
	serialNumber string
	platform     string
	testingDir   string = filepath.Join(`D:\Projects\_work\_pocs\gsim-web-launch\_vendor`)
	gsCli        *cli.Cli
	rootCmd      *cobra.Command
)

func newRootCommand(cli *cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gsim-web-launch",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			p := robotics.Platform(platform)
			winMowerPath, err := downloadAndUnpackWinMower(p, cli)
			if err != nil {
				log.Printf("Failed to download and unpack WinMower: %s", err)
				log.Printf("Closing in 15 seconds...")
				time.Sleep(15 * time.Second)
				return
			}

			gspPaths, err := downloadAndUnpackGSP(serialNumber, platform)
			if err != nil {
				log.Fatalf("Failed to download and unpack GSP: %s", err)
			}
			log.Printf("\nMap: %s\nTestBundle: %s\n", gspPaths.Map, gspPaths.TestBundle)

			fmt.Println("Launching winmower...")
			pkg.LaunchWinMower(winMowerPath, platform)
			time.Sleep(5 * time.Second)

			fmt.Println("Launching test bundle...")
			pkg.RunTestBundle(gspPaths.TestBundle)
			time.Sleep(10 * time.Second)

			fmt.Println("Launching simulator...")
			pkg.LaunchSimulator(gspPaths.Map)
			time.Sleep(5 * time.Second)
		},
	}
	cmd.AddCommand(registry.RegistryCmd)

	cmd.Flags().StringVarP(&serialNumber, "serial-number", "s", "", "Serial number of the device")
	cmd.MarkFlagRequired("serial-number")

	cmd.Flags().StringVarP(&platform, "platform", "p", "P25", "Platform of the device")
	cmd.MarkFlagRequired("platform")
	return cmd
}

func init() {
	cobra.MousetrapHelpText = ""
}

func init() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("Failed to get user cache dir: %s", err)
	}
	wmrCacheDir := filepath.Join(cacheDir, "gsim/winmower")
	err = os.MkdirAll(wmrCacheDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create winmower dir: %s", err)
	}

	bRegistry := robotics.NewBundleRegistry("https://hqvrobotics.azure-api.net")

	gsCli = &cli.Cli{
		AppCacheDir:      cacheDir,
		WinMowerRegistry: robotics.NewWinMowerRegistry(wmrCacheDir, bRegistry),
		BundleRegistry:   bRegistry,
	}

	rootCmd = newRootCommand(gsCli)
}

func Execute(args []string) {
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func downloadAndUnpackWinMower(platform robotics.Platform, gsCli *cli.Cli) (string, error) {
	winMower, err := gsCli.WinMowerRegistry.GetWinMower(platform, context.Background())
	if err != nil {
		return "", err
	}
	if winMower != nil {
		return winMower.Path, nil
	}

	return "", fmt.Errorf("no winmower found")

	// winMower, err := gsCli.WinMowerRegistry.GetCachedWinMower(platform, context.Background())
	// if err != nil || winMower == nil {
	// 	log.Printf("Failed to get cached WinMower: %s", err)
	// } else {
	// 	log.Printf("Found cached WinMower at %s", winMower.Path)
	// 	return winMower.Path, nil
	// }

	// types, err := gsCli.BundleRegistry.FetchBundleTypes(context.Background())
	// if err != nil {
	// 	return "", err
	// }
	// log.Printf("Found %d bundle types\n", len(types))

	// types = robotics.FilterBundleTypes(types, platform)
	// if len(types) == 0 {
	// 	return "", fmt.Errorf("no bundle types found for platform %s", platform)
	// }
	// log.Printf("Found %d bundle types for platform %s\n", len(types), platform)

	// latestType := types[0]
	// log.Printf("Latest bundle type: %s\n", latestType.Name)

	// latestBuild, err := gsCli.BundleRegistry.FetchLatestRelease(context.Background(), latestType.Name)
	// log.Printf("Latest build: %s\n", latestBuild.BlobUrl)
	// if err != nil {
	// 	return "", err
	// }

	// dir := filepath.Join(gsCli.AppCacheDir, "gsim/winmower", platform.String())
	// log.Printf("Downloading and unpacking WinMower to %s", dir)
	// err = pkg.DownloadAndUnpack(latestBuild.BlobUrl, dir)
	// log.Printf("Downloaded and unpacked WinMower to %s", dir)
	// if err != nil {
	// 	log.Printf("Failed to download and unpack WinMower: %s", err)
	// 	return "", err
	// }

	// var exePath string
	// err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
	// 	if filepath.Ext(path) == ".exe" {
	// 		exePath = path
	// 		return nil
	// 	}
	// 	return nil
	// })
	// if exePath == "" {
	// 	return "", fmt.Errorf("no exe found in %s", dir)
	// }
	// log.Printf("Found WinMower exe at %s", exePath)

	// return exePath, err
}

func downloadAndUnpackGSP(serialNumber, platform string) (*robotics.GSPPaths, error) {
	path := filepath.Join(testingDir, "gsp", "GSP_"+serialNumber)
	baseUrl := os.Getenv("GSP_API")
	endpoint := fmt.Sprintf("%s/packet/%s/%s", baseUrl, serialNumber, platform)
	req, err := http.NewRequest("GET", endpoint, nil)
	robotics.AddTifAuthHeaders(req)
	if err != nil {
		return nil, err
	}
	err = ext.DownloadAndUnpack(req, path)
	if err != nil {
		return nil, err
	}

	gsp, err := robotics.LocateGSPPaths(path, serialNumber)
	if err != nil {
		return nil, err
	}

	return gsp, nil
}
