package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/philleif/pomme/internal/blocker"
	"github.com/philleif/pomme/internal/sparkline"
	"github.com/philleif/pomme/internal/storage"
	"github.com/philleif/pomme/internal/timer"
)

const DailyGoal = 12

type Command struct {
	Action string `json:"action"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type StatusData struct {
	TimerState       string `json:"timer_state"`
	Phase            string `json:"phase"`
	Remaining        string `json:"remaining"`
	RemainingSeconds int    `json:"remaining_seconds"`
	IntervalsToday   int    `json:"intervals_today"`
	DailyGoal        int    `json:"daily_goal"`
	BlockEnabled     bool   `json:"block_enabled"`
	AlwaysBlock      bool   `json:"always_block"`
	Sparkline        string `json:"sparkline"`
	StatusLine       string `json:"status_line"`
	WeekValues       []int  `json:"week_values"`
}

type Daemon struct {
	mu       sync.RWMutex
	timer    *timer.Timer
	storage  *storage.Storage
	blocker  *blocker.Blocker
	listener net.Listener

	onStatusChange func(StatusData)
	stopChan       chan struct{}
}

func New() (*Daemon, error) {
	store, err := storage.New()
	if err != nil {
		return nil, fmt.Errorf("failed to open storage: %w", err)
	}

	t := timer.New(timer.DefaultConfig())
	b := blocker.New()

	d := &Daemon{
		timer:   t,
		storage: store,
		blocker: b,
	}

	todayCount, _ := store.TodayCount()
	t.SetIntervalsToday(todayCount)

	t.SetOnComplete(d.onPhaseComplete)

	return d, nil
}

func (d *Daemon) onPhaseComplete(phase timer.Phase) {
	if phase == timer.PhaseWork {
		d.storage.RecordInterval()
		d.sendNotification("Work interval complete!", "Time for a break.")
	} else {
		d.sendNotification("Break complete!", "Ready to focus?")
	}

	d.blocker.SetInInterval(d.timer.Phase() == timer.PhaseWork && d.timer.State() == timer.StateRunning)
	d.notifyStatusChange()
}

func (d *Daemon) sendNotification(title, message string) {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	exec.Command("osascript", "-e", script).Run()
}

func (d *Daemon) SetOnStatusChange(fn func(StatusData)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onStatusChange = fn
}

func (d *Daemon) notifyStatusChange() {
	d.mu.RLock()
	fn := d.onStatusChange
	d.mu.RUnlock()

	if fn != nil {
		fn(d.GetStatus())
	}
}

func (d *Daemon) Start(socketPath string) error {
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	d.listener = listener

	d.blocker.SetEnabled(true)
	d.blocker.Start()

	d.stopChan = make(chan struct{})
	go d.acceptConnections()
	go d.statusUpdateLoop()

	return nil
}

func (d *Daemon) statusUpdateLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			if d.timer.State() == timer.StateRunning {
				d.notifyStatusChange()
			}
		}
	}
}

func (d *Daemon) Stop() {
	if d.stopChan != nil {
		close(d.stopChan)
	}
	if d.listener != nil {
		d.listener.Close()
	}
	d.blocker.Stop()
	d.storage.Close()
}

func (d *Daemon) acceptConnections() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.stopChan:
				return
			default:
				continue
			}
		}
		go d.handleConnection(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	var cmd Command
	if err := json.Unmarshal([]byte(line), &cmd); err != nil {
		d.sendResponse(conn, Response{Success: false, Error: "invalid command"})
		return
	}

	resp := d.handleCommand(cmd)
	d.sendResponse(conn, resp)
}

func (d *Daemon) handleCommand(cmd Command) Response {
	switch cmd.Action {
	case "status":
		return Response{Success: true, Data: d.GetStatus()}

	case "start":
		if d.timer.State() == timer.StatePaused {
			d.timer.Resume()
		} else {
			d.timer.Start()
		}
		d.blocker.SetInInterval(d.timer.Phase() == timer.PhaseWork)
		d.notifyStatusChange()
		return Response{Success: true, Data: d.GetStatus()}

	case "pause":
		d.timer.Pause()
		d.blocker.SetInInterval(false)
		d.notifyStatusChange()
		return Response{Success: true, Data: d.GetStatus()}

	case "skip":
		d.timer.Skip()
		d.blocker.SetInInterval(false)
		d.notifyStatusChange()
		return Response{Success: true, Data: d.GetStatus()}

	case "reset":
		d.timer.Reset()
		d.blocker.SetInInterval(false)
		d.notifyStatusChange()
		return Response{Success: true, Data: d.GetStatus()}

	case "toggle_block":
		d.blocker.ToggleEnabled()
		d.notifyStatusChange()
		return Response{Success: true, Data: d.GetStatus()}

	case "toggle_always":
		d.blocker.ToggleAlwaysBlock()
		d.notifyStatusChange()
		return Response{Success: true, Data: d.GetStatus()}

	default:
		return Response{Success: false, Error: "unknown action"}
	}
}

func (d *Daemon) sendResponse(conn net.Conn, resp Response) {
	data, _ := json.Marshal(resp)
	conn.Write(append(data, '\n'))
}

func (d *Daemon) GetStatus() StatusData {
	status := d.timer.Status()

	days, _ := d.storage.Last7Days()
	intervals := make([]int, len(days))
	for i, day := range days {
		intervals[i] = day.Intervals
	}
	spark := sparkline.GenerateBrailleSpaced(intervals, DailyGoal)

	remaining := status.Remaining
	if remaining < 0 {
		remaining = 0
	}
	mins := int(remaining.Minutes())
	secs := int(remaining.Seconds()) % 60

	var icon string
	switch status.Phase {
	case timer.PhaseWork:
		icon = "🍅"
	case timer.PhaseShortBreak, timer.PhaseLongBreak:
		icon = "☕"
	}

	if status.State == timer.StatePaused {
		icon = "⏸"
	}

	statusLine := fmt.Sprintf("%s %02d:%02d %s", icon, mins, secs, spark)

	// Enhanced status line with subscript for today's count
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)
	statusLine = sparkline.CompactStatus(icon, timeStr, spark, status.IntervalsToday)

	return StatusData{
		TimerState:       status.State.String(),
		Phase:            status.Phase.String(),
		Remaining:        fmt.Sprintf("%02d:%02d", mins, secs),
		RemainingSeconds: int(remaining.Seconds()),
		IntervalsToday:   status.IntervalsToday,
		DailyGoal:        DailyGoal,
		BlockEnabled:     d.blocker.Enabled(),
		AlwaysBlock:      d.blocker.AlwaysBlock(),
		Sparkline:        spark,
		StatusLine:       statusLine,
		WeekValues:       intervals,
	}
}

func (d *Daemon) Timer() *timer.Timer {
	return d.timer
}

func (d *Daemon) Blocker() *blocker.Blocker {
	return d.blocker
}
