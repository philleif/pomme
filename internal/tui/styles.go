package tui

import "github.com/charmbracelet/lipgloss"

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6347")).
			Align(lipgloss.Center)

	timerStyleNormal = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#FF6347"))

	counterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	timerBreakStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#4A90D9"))

	timerPausedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#888888"))

	phaseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Align(lipgloss.Center)

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

	progressFullStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6347"))

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333"))

	sparklineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73F59F"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	goalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	todayValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6347")).
			Bold(true)

	toggleOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73F59F")).
			Bold(true)

	toggleOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6347"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6347")).
			Padding(1, 2)
)
