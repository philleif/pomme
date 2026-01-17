package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/philleif/pomme/internal/client"
	"github.com/philleif/pomme/internal/daemon"
)

type tickMsg time.Time

type Model struct {
	client *client.Client
	status *daemon.StatusData
	err    error
	width  int
	height int
}

func NewModel() Model {
	c := client.New()
	status, err := c.Status()

	return Model{
		client: c,
		status: status,
		err:    err,
		width:  60,
		height: 20,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		status, err := m.client.Status()
		m.status = status
		m.err = err
		return m, tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "s":
			m.client.Start()
			status, _ := m.client.Status()
			m.status = status
			return m, nil

		case "p":
			m.client.Pause()
			status, _ := m.client.Status()
			m.status = status
			return m, nil

		case "k":
			m.client.Skip()
			status, _ := m.client.Status()
			m.status = status
			return m, nil

		case "r":
			m.client.Reset()
			status, _ := m.client.Status()
			m.status = status
			return m, nil

		case "b":
			m.client.ToggleBlock()
			status, _ := m.client.Status()
			m.status = status
			return m, nil

		case "a":
			m.client.ToggleAlways()
			status, _ := m.client.Status()
			m.status = status
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return boxStyle.Render(fmt.Sprintf("Error: %v\n\nMake sure daemon is running:\n  pomme --daemon", m.err))
	}

	if m.status == nil {
		return boxStyle.Render("Connecting...")
	}

	var b strings.Builder

	// Determine state color: bright green for work, pale dark green for break, red for inactive
	var stateColor lipgloss.Color
	switch {
	case m.status.TimerState == "running" && m.status.Phase == "work":
		stateColor = colorWorkActive
	case m.status.TimerState == "running":
		stateColor = colorBreakActive
	default:
		stateColor = colorInactive
	}

	// Header line: title | timer | counter (space-between)
	timerFg := lipgloss.Color("#FFFFFF")
	if m.status.TimerState == "running" {
		timerFg = lipgloss.Color("#000000")
	}
	timerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(timerFg).
		Background(stateColor)

	title := titleStyle.Render("🍅 POMME")
	timer := timerStyle.Render(fmt.Sprintf(" %s ", m.status.Remaining))
	counter := counterStyle.Render(fmt.Sprintf("%d/%d", m.status.IntervalsToday, m.status.DailyGoal))

	headerWidth := 38
	spacer := headerWidth - lipgloss.Width(title) - lipgloss.Width(timer) - lipgloss.Width(counter)
	if spacer < 2 {
		spacer = 2
	}
	leftSpace := strings.Repeat(" ", spacer/2)
	rightSpace := strings.Repeat(" ", spacer-spacer/2)

	b.WriteString(title + leftSpace + timer + rightSpace + counter)
	b.WriteString("\n\n")

	help := helpStyle.Render("[s]tart  [p]ause  [k]ip  [r]eset")
	b.WriteString(help)
	b.WriteString("\n\n")

	// Enhanced progress display with goal reference
	progress := m.renderProgress(m.status.IntervalsToday, m.status.DailyGoal)
	b.WriteString(statsStyle.Render(fmt.Sprintf("Today: %s %d", progress, m.status.IntervalsToday)))
	b.WriteString("\n")
	b.WriteString(goalStyle.Render(fmt.Sprintf("       %s goal:%d", strings.Repeat("─", 12), m.status.DailyGoal)))
	b.WriteString("\n\n")

	// Enhanced sparkline with day labels and values
	// Use fixed-width formatting for alignment (3 chars per column)
	b.WriteString(statsStyle.Render("Week:  "))
	b.WriteString(sparklineStyle.Render(m.status.Sparkline))
	b.WriteString("\n")
	
	// Day labels - dynamically generated based on today's day of week
	// Last 7 days ending with today
	dayNames := []string{"S", "M", "T", "W", "T", "F", "S"}
	today := int(time.Now().Weekday()) // 0=Sunday, 1=Monday, etc.
	var dayLabels strings.Builder
	dayLabels.WriteString("       ")
	for i := 0; i < 7; i++ {
		// 6 days ago, 5 days ago, ..., today
		dayIdx := (today - 6 + i + 7) % 7
		dayLabels.WriteString(fmt.Sprintf("%-3s", dayNames[dayIdx]))
	}
	b.WriteString(labelStyle.Render(dayLabels.String()))
	b.WriteString("\n")
	
	// Show interval counts - 3 char width per column
	if len(m.status.WeekValues) > 0 {
		b.WriteString(labelStyle.Render("       "))
		for i, v := range m.status.WeekValues {
			if i == len(m.status.WeekValues)-1 {
				// Highlight today
				b.WriteString(todayValueStyle.Render(fmt.Sprintf("%-3d", v)))
			} else {
				b.WriteString(labelStyle.Render(fmt.Sprintf("%-3d", v)))
			}
		}
		b.WriteString("\n")
		
		// Today marker - aligned under last column
		todayPos := len(m.status.WeekValues) - 1
		b.WriteString(labelStyle.Render(strings.Repeat(" ", 7+todayPos*3) + "↑today"))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	blockStatus := m.renderToggle("Block Messages", m.status.BlockEnabled, "b")
	b.WriteString(blockStatus)
	b.WriteString("\n")

	alwaysStatus := m.renderToggle("Always block", m.status.AlwaysBlock, "a")
	b.WriteString(alwaysStatus)
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("[q]uit"))

	// Use state-colored border
	dynamicBox := boxStyle.BorderForeground(stateColor)
	content := dynamicBox.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderProgress(current, total int) string {
	width := 12
	filled := (current * width) / total
	if filled > width {
		filled = width
	}

	full := progressFullStyle.Render(strings.Repeat("█", filled))
	empty := progressEmptyStyle.Render(strings.Repeat("░", width-filled))

	return full + empty
}

func (m Model) renderToggle(label string, on bool, key string) string {
	var status string
	if on {
		status = toggleOnStyle.Render("ON")
	} else {
		status = toggleOffStyle.Render("OFF")
	}
	return fmt.Sprintf("[%s] %s: %s", key, label, status)
}

func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
