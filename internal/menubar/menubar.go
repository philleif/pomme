package menubar

import (
	"os/exec"

	"fyne.io/systray"
	"github.com/philleif/pomme/internal/daemon"
)

type MenuBar struct {
	daemon *daemon.Daemon

	mStart      *systray.MenuItem
	mPause      *systray.MenuItem
	mSkip       *systray.MenuItem
	mReset      *systray.MenuItem
	mBlock      *systray.MenuItem
	mAlways     *systray.MenuItem
	mOpenTUI    *systray.MenuItem
	mQuit       *systray.MenuItem
}

func New(d *daemon.Daemon) *MenuBar {
	return &MenuBar{daemon: d}
}

func (m *MenuBar) Run() {
	systray.Run(m.onReady, m.onExit)
}

func (m *MenuBar) onReady() {
	systray.SetTitle("🍅 25:00")
	systray.SetTooltip("Pomme - Pomodoro Timer")

	m.mStart = systray.AddMenuItem("Start", "Start timer")
	m.mPause = systray.AddMenuItem("Pause", "Pause timer")
	m.mSkip = systray.AddMenuItem("Skip", "Skip to next phase")
	m.mReset = systray.AddMenuItem("Reset", "Reset timer")

	systray.AddSeparator()

	m.mBlock = systray.AddMenuItemCheckbox("Block Messages", "Block Messages during focus", m.daemon.Blocker().Enabled())
	m.mAlways = systray.AddMenuItemCheckbox("Always Block", "Block Messages even between intervals", m.daemon.Blocker().AlwaysBlock())

	systray.AddSeparator()

	m.mOpenTUI = systray.AddMenuItem("Open Controls", "Open TUI in terminal")

	systray.AddSeparator()

	m.mQuit = systray.AddMenuItem("Quit", "Quit Pomme")

	m.daemon.SetOnStatusChange(m.updateStatus)

	go m.handleClicks()
}

func (m *MenuBar) handleClicks() {
	for {
		select {
		case <-m.mStart.ClickedCh:
			m.daemon.Timer().Start()
			m.daemon.Blocker().SetInInterval(true)

		case <-m.mPause.ClickedCh:
			m.daemon.Timer().Pause()
			m.daemon.Blocker().SetInInterval(false)

		case <-m.mSkip.ClickedCh:
			m.daemon.Timer().Skip()

		case <-m.mReset.ClickedCh:
			m.daemon.Timer().Reset()
			m.daemon.Blocker().SetInInterval(false)

		case <-m.mBlock.ClickedCh:
			enabled := m.daemon.Blocker().ToggleEnabled()
			if enabled {
				m.mBlock.Check()
			} else {
				m.mBlock.Uncheck()
			}

		case <-m.mAlways.ClickedCh:
			always := m.daemon.Blocker().ToggleAlwaysBlock()
			if always {
				m.mAlways.Check()
			} else {
				m.mAlways.Uncheck()
			}

		case <-m.mOpenTUI.ClickedCh:
			m.openTUI()

		case <-m.mQuit.ClickedCh:
			systray.Quit()
		}
	}
}

func (m *MenuBar) openTUI() {
	script := `tell application "Terminal"
		activate
		do script "pomme"
	end tell`
	exec.Command("osascript", "-e", script).Run()
}

func (m *MenuBar) updateStatus(status daemon.StatusData) {
	systray.SetTitle(status.StatusLine)

	if m.daemon.Blocker().Enabled() {
		m.mBlock.Check()
	} else {
		m.mBlock.Uncheck()
	}

	if m.daemon.Blocker().AlwaysBlock() {
		m.mAlways.Check()
	} else {
		m.mAlways.Uncheck()
	}
}

func (m *MenuBar) onExit() {
	m.daemon.Stop()
}
