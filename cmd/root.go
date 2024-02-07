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

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
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
	teaApp := tea.NewProgram(initialModel(resChan, errChan), tea.WithAltScreen())
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
	err = testRunner.Run(context.Background(), `D:\Projects\_work\GardenTVAutoLoader\GardenTVAutoloader\Resources\testscript.zip`, "-tcpAddress", "127.0.0.1:4250")
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

type model struct {
	resChan  chan runtimeConfig
	errChan  chan error
	msgChan  chan progressMsg
	text     string
	progress progress.Model
	spinner  spinner.Model
	width    int
	height   int
}

func initialModel(resChan chan runtimeConfig, errChan chan error) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#8b5cf6"))
	return model{
		errChan: errChan,
		resChan: resChan,
		msgChan: make(chan progressMsg),
		text:    "",
		progress: progress.New(
			progress.WithWidth(40),
			progress.WithDefaultGradient(),
			progress.WithoutPercentage(),
			progress.WithSpringOptions(10, 1),
		),
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		prepareRuntime(m.msgChan, m.resChan, m.errChan),
		receiveProgressMsg(m.msgChan),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.errChan <- fmt.Errorf("User quit")
			return m, tea.Quit

		default:
			return m, nil
		}

	case progressMsg:
		pm := progressMsg(msg)

		if pm.isError {
			m.text = fmt.Sprintf("An error occurred\n%s ... Q to quit", pm.text)
			progressCmd := m.progress.SetPercent(0.0)
			return m, progressCmd
		}

		m.text = string(fmt.Sprintf("%s", pm.text))
		progressCmd := m.progress.SetPercent(float64(pm.percent) / 100.0)
		return m, tea.Batch(
			receiveProgressMsg(m.msgChan),
			progressCmd,
			m.spinner.Tick,
		)

	case completeMsg:
		return m, tea.Quit

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, tea.Batch(cmd, m.spinner.Tick)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.spinner.Tick)
	}

	return m, nil
}

func (m model) View() string {
	m.progress.Width = m.width / 4
	prog := m.progress.View()
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				Margin(1, 0).
				Foreground(lipgloss.Color("#ffffff")).
				Render("Garden Simulator Launcher"),
			lipgloss.NewStyle().
				MarginBottom(1).
				Foreground(lipgloss.Color("#aaaaaa")).
				Render(m.spinner.View()+m.text),
			prog,
		),
	)
}

type progressMsg struct {
	text    string
	percent int
	isError bool
}
type completeMsg struct{}

func prepareRuntime(msgChan chan progressMsg, resChan chan runtimeConfig, errChan chan error) tea.Cmd {
	return func() tea.Msg {
		msgChan <- progressMsg{text: fmt.Sprintf("Downloading and unpacking %s winmower...", platform), percent: 30}
		p := robotics.Platform(platform)
		winMower, err := gsCli.WinMowerRegistry.GetWinMower(p, context.Background())
		if err != nil {
			msgChan <- progressMsg{text: "Failed to get winmower", isError: true}
			errChan <- err
			return nil
		}
		if winMower == nil {
			msgChan <- progressMsg{text: fmt.Sprintf("No winmower found for platform %s", platform), isError: true}
			errChan <- fmt.Errorf("no winmower found for platform %s", platform)
			return nil
		}

		msgChan <- progressMsg{text: "Fetching the Garden Simulator Packet...", percent: 60}
		gspPaths, err := gsCli.GSPRegistry.GetGSP(serialNumber, platform)
		if err != nil {
			msgChan <- progressMsg{text: fmt.Sprintf("Failed to download and unpack GSP: %s", err), isError: true}
			errChan <- err
			return nil
		}

		msgChan <- progressMsg{text: "Downloading and unpacking Garden Simulator...", percent: 100}
		simulator, err := gsCli.SimulatorRegistry.GetSimulator(context.Background())
		if err != nil {
			msgChan <- progressMsg{text: fmt.Sprintf("Failed to get simulator: %s", err), isError: true}
			errChan <- err
			return nil
		}

		resChan <- runtimeConfig{
			Winmower:  winMower,
			Simulator: simulator,
			GSPPaths:  gspPaths,
		}
		return completeMsg{}
	}
}

func receiveProgressMsg(mshChan chan progressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-mshChan
	}
}
