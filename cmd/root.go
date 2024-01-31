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
	testingDir   string
)

var rootCmd = &cobra.Command{
	Use:   "gsim-web-launch",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		gspPaths, err := downloadAndUnpackGSP(serialNumber, platform)
		if err != nil {
			log.Fatalf("Failed to download and unpack GSP: %s", err)
		}

		log.Printf("\nMap: %s\nTestBundle: %s\n", gspPaths.Map, gspPaths.TestBundle)

		fmt.Println("Launching winmower...")
		pkg.LaunchWinMower()
		time.Sleep(5 * time.Second)
		fmt.Println("Launching simulator...")
		pkg.LaunchSimulator(gspPaths.Map)
		time.Sleep(5 * time.Second)
		fmt.Println("Launching test bundle...")
		pkg.RunTestBundle(gspPaths.TestBundle)
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
