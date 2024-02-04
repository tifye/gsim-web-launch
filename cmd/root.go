package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Tifufu/gsim-web-launch/cmd/clear"
	"github.com/Tifufu/gsim-web-launch/cmd/cli"
	"github.com/Tifufu/gsim-web-launch/cmd/registry"
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
			runRootCommand(cli)
		},
	}
	cmd.Flags().StringVarP(&serialNumber, "serial-number", "s", "", "Serial number of the device")
	cmd.MarkFlagRequired("serial-number")

	cmd.Flags().StringVarP(&platform, "platform", "p", "P25", "Platform of the device")
	cmd.MarkFlagRequired("platform")

	cmd.AddCommand(
		registry.RegistryCmd,
		clear.NewClearCommand(cli),
	)
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
	appCacheDir := filepath.Join(cacheDir, "gsim")

	wmrCacheDir := filepath.Join(appCacheDir, "winmower")
	err = os.MkdirAll(wmrCacheDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create winmower dir: %s", err)
	}

	bRegistry := robotics.NewBundleRegistry("https://hqvrobotics.azure-api.net")
	gsCli = &cli.Cli{
		AppCacheDir:      appCacheDir,
		WinMowerRegistry: robotics.NewWinMowerRegistry(wmrCacheDir, bRegistry),
		BundleRegistry:   bRegistry,
		GSPRegistry:      robotics.NewGSPRegistry(filepath.Join(appCacheDir, "gsp"), os.Getenv("GSP_API")),
	}

	rootCmd = newRootCommand(gsCli)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runRootCommand(cli *cli.Cli) {
	defer func() {
		if r := recover(); r != nil {
			log.Info("Recovered in root command", r)
			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')
			return
		}
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
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

	wmRunner, err := createWinMowerRunner(cli.AppCacheDir, winMower)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("Starting winmower...")
	err = wmRunner.Start(context.Background())
	if err != nil {
		log.Error("Failed to start winmower", "err", err)
		return
	}
	defer wmRunner.Stop()
	time.Sleep(3 * time.Second)

	log.Info("Running test bundle...")
	testRunner := createTestBundleRunner()
	err = testRunner.Run(context.Background(), gspPaths.TestBundle, "-tcpAddress", "127.0.0.1:4250")
	if err != nil {
		log.Error("Failed to start test bundle", "err", err)
		return
	}

	log.Info("Launching simulator...")
	simPath := os.Getenv("SIM_PATH")
	if simPath == "" {
		log.Error("SIM_PATH not set")
		return
	}
	err = runner.LaunchSimulator(simPath, gspPaths.Map)
	if err != nil {
		log.Error("Failed to launch simulator", "err", err)
		return
	}

	time.Sleep(5 * time.Second)
	err = testRunner.Run(context.Background(), `D:\Projects\_work\GardenTVAutoLoader\GardenTVAutoloader\Resources\testscript.zip`, "-tcpAddress", "127.0.0.1:4250")
	if err != nil {
		log.Error("Failed to start test bundle", "err", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func createTestBundleRunner() *runner.TestBundleRunner {
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
		Prefix:          "TifConsole.Auto",
	})
	tifColor := lipgloss.Color("#3b82f6")
	style := log.DefaultStyles()
	style.Prefix = lipgloss.NewStyle().Foreground(tifColor)
	style.Timestamp = lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeftBackground(tifColor).
		BorderLeftForeground(tifColor).
		BorderLeft(true).
		PaddingLeft(1)
	logger.SetStyles(style)
	testLogger := runner.NewTifConsoleLogger(logger)
	testRunner := runner.NewTestBundleRunner(`C:\Users\demat\AppData\Local\TifApp\TifConsole.Auto.exe`, testLogger)
	return testRunner
}

func createWinMowerRunner(cacheDir string, winMower *robotics.WinMower) (*runner.WinMowerRunner, error) {
	wmDir := filepath.Join(cacheDir, "winmower-filesystems", platform)
	err := os.MkdirAll(wmDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create winmower dir: %w", err)
	}
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
		Prefix:          "WinMower",
	})
	wmColor := lipgloss.Color("#8b5cf6")
	style := log.DefaultStyles()
	style.Prefix = lipgloss.NewStyle().Foreground(wmColor)
	style.Timestamp = lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeftBackground(wmColor).
		BorderLeftForeground(wmColor).
		BorderLeft(true).
		PaddingLeft(1)
	logger.SetStyles(style)
	wmLogger := runner.NewWinMowerLogger(logger)
	wmRunner := runner.NewWinMowerRunner(wmDir, winMower.Path, wmLogger)
	return wmRunner, nil
}
