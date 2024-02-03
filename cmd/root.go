package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd/cli"
	"github.com/Tifufu/gsim-web-launch/cmd/registry"
	"github.com/Tifufu/gsim-web-launch/pkg"
	"github.com/Tifufu/gsim-web-launch/pkg/robotics"
	"github.com/Tifufu/gsim-web-launch/pkg/runner"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/lipgloss"
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
					log.Info("Recovered in root command", r)
					reader := bufio.NewReader(os.Stdin)
					reader.ReadString('\n')
				}
			}()

			log.Info("Getting winmower...")
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

			wmDir := filepath.Join(cli.AppCacheDir, "gsim/winmower-filesystems", platform)
			err = os.MkdirAll(wmDir, 0755)
			if err != nil {
				log.Error("Failed to create winmower dir", "err", err)
				return
			}

			log.Info("Launching winmower...")
			wmSubLogger := log.NewWithOptions(os.Stdout, log.Options{
				ReportCaller:    false,
				ReportTimestamp: false,
				Prefix:          "WinMower\t",
			})
			style := log.DefaultStyles()
			style.Prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#8b5cf6"))
			wmSubLogger.SetStyles(style)
			wmLogger := runner.NewWinMowerLogger(wmSubLogger)
			wmRunner := runner.NewWinMowerRunner(wmDir, winMower.Path, wmLogger)
			err = wmRunner.Start(context.Background())
			if err != nil {
				log.Error("Failed to start winmower", "err", err)
				return
			}
			defer wmRunner.Stop()
			//pkg.LaunchWinMower(winMower.Path, platform)
			time.Sleep(5 * time.Second)

			log.Info("Running test bundle...")
			testSubLogger := log.NewWithOptions(os.Stdout, log.Options{
				ReportCaller:    false,
				ReportTimestamp: false,
				Prefix:          "TifConsole.Auto\t",
			})
			style = log.DefaultStyles()
			style.Prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#f43f5e"))
			testSubLogger.SetStyles(style)
			testLogger := runner.NewTifConsoleLogger(testSubLogger)
			testRunner := runner.NewTestBundleRunner(`C:\Users\demat\AppData\Local\TifApp\TifConsole.Auto.exe`, testLogger)
			err = testRunner.Run(context.Background(), gspPaths.TestBundle, "-tcpAddress", "127.0.0.1:4250")
			if err != nil {
				log.Error("Failed to start test bundle", "err", err)
				return
			}
			//pkg.RunTestBundle(gspPaths.TestBundle)

			log.Info("Launching simulator...")
			pkg.LaunchSimulator(gspPaths.Map)

			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')
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
