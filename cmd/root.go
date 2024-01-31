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
)

var rootCmd = &cobra.Command{
	Use:   "gsim-web-launch",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		testingDir := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor`

		gspZipDir := filepath.Join(testingDir, "gsp_zips")
		err := os.MkdirAll(gspZipDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create GSP zip directory: %s", err)
		}

		gspZipPath := filepath.Join(gspZipDir, "GSP_"+serialNumber+".zip")
		err = pkg.DownloadGSP(serialNumber, platform, gspZipPath)
		if err != nil {
			log.Fatalf("Failed to download GSP: %s", err)
		}

		unzipDest := fmt.Sprintf("%s\\gsp_unzips\\GSP_%s", testingDir, serialNumber)
		err = pkg.Unzip(gspZipPath, unzipDest)
		if err != nil {
			log.Fatalf("Failed to unzip GSP: %s", err)
		}
		log.Printf("Unzipped GSP to %s\n", unzipDest)

		gspPaths, err := pkg.LocateGSPPaths(unzipDest, serialNumber)
		if err != nil {
			log.Fatalf("Failed to locate GSP paths: %s", err)
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
