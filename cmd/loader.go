package cmd

import (
	"context"
	"fmt"

	"github.com/Tifufu/gsim-web-launch/pkg/robotics"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
		msgChan: make(chan progressMsg, 1),
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

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		if m.progress.Percent() > 0.99 && !m.progress.IsAnimating() {
			return m, tea.Quit
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
	m.progress.Width = m.width / 3
	prog := m.progress.View()

	block := lipgloss.JoinVertical(
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
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8b5cf6")).
		Padding(1).
		MarginBottom(2).
		Render(block)
}

type progressMsg struct {
	text    string
	percent int
	isError bool
}

func prepareRuntime(msgChan chan progressMsg, resChan chan runtimeConfig, errChan chan error) tea.Cmd {
	return func() tea.Msg {
		msgChan <- progressMsg{text: fmt.Sprintf("Downloading and unpacking %s winmower...", platform), percent: 0}
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

		msgChan <- progressMsg{text: "Fetching the Garden Simulator Packet...", percent: 30}
		gspPaths, err := gsCli.GSPRegistry.GetGSP(serialNumber, platform)
		if err != nil {
			msgChan <- progressMsg{text: fmt.Sprintf("Failed to download and unpack GSP: %s", err), isError: true}
			errChan <- err
			return nil
		}

		msgChan <- progressMsg{text: "Downloading and unpacking Garden Simulator...", percent: 60}
		simulator, err := gsCli.SimulatorRegistry.GetSimulator(context.Background())
		if err != nil {
			msgChan <- progressMsg{text: fmt.Sprintf("Failed to get simulator: %s", err), isError: true}
			errChan <- err
			return nil
		}

		msgChan <- progressMsg{text: "Preparation complete", percent: 100}

		resChan <- runtimeConfig{
			Winmower:  winMower,
			Simulator: simulator,
			GSPPaths:  gspPaths,
		}

		return nil
	}
}

func receiveProgressMsg(mshChan chan progressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-mshChan
	}
}
