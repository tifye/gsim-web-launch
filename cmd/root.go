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
	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	serialNumber string
	platform     string
	gsCli        *cli.Cli
	rootCmd      *cobra.Command
)

type runtimeConfig struct {
	Winmower  *robotics.WinMower
	Simulator *robotics.Simulator
	GSPPaths  *robotics.GSPPaths
}

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
	initConfig()
	cobra.MousetrapHelpText = ""
}

func Execute(args []string) {
	v := viper.GetViper()
	wmDir := v.GetString("directories.winMowers")
	err := os.MkdirAll(wmDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create winmower dir: %s", err)
	}

	bRegistry := robotics.NewBundleRegistry(v.GetString("endpoints.bundleStorage"))
	gsCli = &cli.Cli{
		Config:            v,
		AppCacheDir:       v.GetString("directories.appCacheDir"),
		BundleRegistry:    bRegistry,
		WinMowerRegistry:  robotics.NewWinMowerRegistry(wmDir, bRegistry),
		SimulatorRegistry: robotics.NewSimulatorRegistry(v.GetString("directories.simulator"), bRegistry),
		GSPRegistry:       robotics.NewGSPRegistry(v.GetString("directories.gardenSimulatorPackets"), v.GetString("endpoints.gardenSimulatorPacket")),
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
			log.Warn("Recovered in root command", r)
			log.Info("Press enter to exit...")
			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')
		}
	}()

	log.SetLevel(log.InfoLevel)

	var resChan = make(chan runtimeConfig, 1)
	var errChan = make(chan error, 1)
	teaApp := tea.NewProgram(initialModel(resChan, errChan))
	if _, err := teaApp.Run(); err != nil {
		log.Error("Failed to start tea program", "err", err)
		return
	}

	var runtime runtimeConfig
	select {
	case runtime = <-resChan:
	case err := <-errChan:
		log.Error("Failed to prepare runtime", "err", err)
		return
	}

	log.SetLevel(log.DebugLevel)

	wmRunner, err := createWinMowerRunner(cli.Config.GetString("directories.winMowerFileSystems"), runtime.Winmower)
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
	testRunner := createTestBundleRunner(gsCli.Config.GetString("programs.tifConsole"))
	err = testRunner.Run(context.Background(), runtime.GSPPaths.TestBundle, "-tcpAddress", "127.0.0.1:4250")
	if err != nil {
		log.Error("Failed to start test bundle", "err", err)
		return
	}

	log.Info("Launching simulator...")
	err = runner.LaunchSimulator(runtime.Simulator.Path, runtime.GSPPaths.Map)
	if err != nil {
		log.Error("Failed to launch simulator", "err", err)
		return
	}

	log.Info("Running start trigger test bundle...")
	err = testRunner.Run(context.Background(), `C:\Repositories\GardenTVAutoLoader\GardenTVAutoloader\Resources\testscript.zip`, "-tcpAddress", "127.0.0.1:4250")
	if err != nil {
		log.Error("Failed to start test bundle", "err", err)
		return
	}

	log.Info("Press enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func createTestBundleRunner(tifConsolePath string) *runner.TestBundleRunner {
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
	testRunner := runner.NewTestBundleRunner(tifConsolePath, testLogger)
	return testRunner
}

func createWinMowerRunner(wmFsCacheDir string, winMower *robotics.WinMower) (*runner.WinMowerRunner, error) {
	wmDir := filepath.Join(wmFsCacheDir, platform)
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
