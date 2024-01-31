package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd/registry"
	"github.com/Tifufu/gsim-web-launch/pkg"
	"github.com/spf13/cobra"
)

var (
	serialNumber string
)

var rootCmd = &cobra.Command{
	Use:   "gsim-web-launch",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		gspZipPath := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GSP_190703524.zip`
		unzipDest := `D:\Projects\_work\_pocs\gsim-web-launch\_vendor\GSP_190703524`
		err := pkg.Unzip(gspZipPath, unzipDest)
		if err != nil {
			log.Fatalf("Failed to unzip GSP: %s", err)
		}
		log.Printf("Unzipped GSP to %s\n", unzipDest)

		gspPaths, err := pkg.LocateGSPPaths(unzipDest, "190703524")
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
}

func Execute(args []string) {
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
