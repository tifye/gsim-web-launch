package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd/cli"
	"github.com/Tifufu/gsim-web-launch/cmd/registry"
	"github.com/Tifufu/gsim-web-launch/pkg"
	"github.com/Tifufu/gsim-web-launch/pkg/robotics"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/log"
)

var (
	serialNumber string
	platform     string
	gsCli        *cli.Cli
	rootCmd      *cobra.Command
)

func newRootCommand(cli *cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gsim-web-launch",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			defer func() {
				if r := recover(); r != nil {
					log.Info("Recovered in main", r)
					time.Sleep(15 * time.Second)
				}
				log.Info("Closing in 15 seconds...")
				time.Sleep(15 * time.Second)
			}()

			p := robotics.Platform(platform)
			winMower, err := gsCli.WinMowerRegistry.GetWinMower(p, context.Background())
			if err != nil {
				log.Error("Failed to get winmower", "err", err)
				return
			}
			if winMower == nil {
				log.Error("No winmower found for platform", "platform", platform)
				return
			}

			gspPaths, err := gsCli.GSPRegistry.GetGSP(serialNumber, platform)
			if err != nil {
				log.Error("Failed to download and unpack GSP", "err", err)
				return
			}

			log.Info("Launching winmower...")
			pkg.LaunchWinMower(winMower.Path, platform)
			time.Sleep(5 * time.Second)

			log.Info("Launching test bundle...")
			pkg.RunTestBundle(gspPaths.TestBundle)

			log.Info("Launching simulator...")
			pkg.LaunchSimulator(gspPaths.Map)
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

func Execute(args []string) {
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
		GSPRegistry:      robotics.NewGSPRegistry(filepath.Join(cacheDir, "gsim/gsp"), os.Getenv("GSP_API")),
	}

	rootCmd = newRootCommand(gsCli)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
