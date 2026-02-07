package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/philleif/pomme/internal/client"
	"github.com/philleif/pomme/internal/config"
	"github.com/philleif/pomme/internal/daemon"
	"github.com/philleif/pomme/internal/mcp"
	"github.com/philleif/pomme/internal/menubar"
	"github.com/philleif/pomme/internal/sparkline"
	"github.com/philleif/pomme/internal/storage"
	"github.com/philleif/pomme/internal/tui"
)

func main() {
	daemonMode := flag.Bool("daemon", false, "Run as daemon (menu bar only)")
	mcpMode := flag.Bool("mcp", false, "Run as MCP server (stdio)")
	statusMode := flag.Bool("status", false, "Print status line (for tmux)")
	simpleBarMode := flag.Bool("simplebar", false, "Print status for simple-bar widget")
	sketchyBarMode := flag.Bool("sketchybar", false, "Print JSON status for SketchyBar widget")
	startCmd := flag.Bool("start", false, "Start/resume timer")
	pauseCmd := flag.Bool("pause", false, "Pause timer")
	skipCmd := flag.Bool("skip", false, "Skip to next phase")
	resetCmd := flag.Bool("reset", false, "Reset timer")
	toggleBlockCmd := flag.Bool("toggle-block", false, "Toggle Messages blocking")
	toggleAlwaysCmd := flag.Bool("toggle-always", false, "Toggle always block Messages mode")
	toggleCmd := flag.Bool("toggle", false, "Toggle timer start/pause")
	statsCmd := flag.Bool("stats", false, "Print today's stats")
	graphCmd := flag.Bool("graph", false, "Show graphical sparkline (Kitty protocol for Ghostty)")

	flag.Parse()

	c := client.New()

	switch {
	case *daemonMode:
		runDaemon()

	case *mcpMode:
		ensureDaemon(c, true)
		if err := mcp.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}

	case *statusMode:
		status, err := c.Status()
		if err != nil {
			fmt.Println("⏹ --:--")
			os.Exit(0)
		}
		fmt.Println(status.StatusLine)

	case *simpleBarMode:
		status, err := c.Status()
		if err != nil {
			fmt.Println("⏹ --:--")
			os.Exit(0)
		}
		// Compact format for simple-bar: icon time count
		var icon string
		switch {
		case status.TimerState == "paused":
			icon = "⏸"
		case status.TimerState == "idle":
			icon = "⏹"
		case status.Phase == "work":
			icon = "🍅"
		default:
			icon = "☕"
		}
		fmt.Printf("%s %s %s\n", icon, status.Remaining, subscript(status.IntervalsToday))

	case *sketchyBarMode:
		status, err := c.Status()
		if err != nil {
			// Output empty JSON for no daemon
			fmt.Println(`{"drawing":"off"}`)
			os.Exit(0)
		}
		// JSON output for SketchyBar with colors and click actions
		var icon, color string
		switch {
		case status.TimerState == "paused":
			icon = "󰏤"
			color = "0xfff7768e" // red for paused
		case status.TimerState == "idle":
			icon = "󰓛"
			color = "0xff565f89" // grey for idle
		case status.Phase == "work":
			icon = "󰔟"
			color = "0xff9ece6a" // green for work
		default: // break
			icon = "󰒲"
			color = "0xff7dcfff" // cyan for break
		}
		// Block indicator: green dot if always_block enabled, red dot if disabled
		blockIndicator := "●"
		blockColor := "0xfff7768e" // red - not blocking
		if status.AlwaysBlock {
			blockColor = "0xff9ece6a" // green - blocking
		}
		fmt.Printf(`{"icon":"%s","label":"%s %s","color":"%s","block_icon":"%s","block_color":"%s","drawing":"on"}%s`,
			icon, status.Remaining, subscript(status.IntervalsToday), color, blockIndicator, blockColor, "\n")

	case *startCmd:
		ensureDaemon(c, false)
		_, err := c.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Timer started")

	case *pauseCmd:
		ensureDaemon(c, false)
		_, err := c.Pause()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Timer paused")

	case *skipCmd:
		ensureDaemon(c, false)
		_, err := c.Skip()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Skipped to next phase")

	case *resetCmd:
		ensureDaemon(c, false)
		_, err := c.Reset()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Timer reset")

	case *toggleBlockCmd:
		ensureDaemon(c, false)
		status, err := c.ToggleBlock()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if status.BlockEnabled {
			fmt.Println("Messages blocking: ON")
		} else {
			fmt.Println("Messages blocking: OFF")
		}

	case *toggleAlwaysCmd:
		ensureDaemon(c, false)
		status, err := c.ToggleAlways()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if status.AlwaysBlock {
			fmt.Println("Always block Messages: ON")
		} else {
			fmt.Println("Always block Messages: OFF")
		}

	case *toggleCmd:
		ensureDaemon(c, false)
		status, err := c.Status()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if status.TimerState == "running" {
			_, err = c.Pause()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Timer paused")
		} else {
			_, err = c.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Timer started")
		}

	case *statsCmd:
		ensureDaemon(c, false)
		status, err := c.Status()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Today: %d/%d intervals\n", status.IntervalsToday, status.DailyGoal)
		fmt.Printf("Week:  %s\n", status.Sparkline)
		// Dynamic day labels based on today
		dayNames := []string{"S", "M", "T", "W", "T", "F", "S"}
		today := int(time.Now().Weekday())
		fmt.Print("       ")
		for i := 0; i < 7; i++ {
			dayIdx := (today - 6 + i + 7) % 7
			fmt.Printf("%-3s", dayNames[dayIdx])
		}
		fmt.Println()
		if len(status.WeekValues) > 0 {
			fmt.Print("       ")
			for _, v := range status.WeekValues {
				fmt.Printf("%-3d", v)
			}
			fmt.Println()
		}

	case *graphCmd:
		ensureDaemon(c, false)
		status, err := c.Status()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Today: %d/%d intervals\n", status.IntervalsToday, status.DailyGoal)
		fmt.Println()
		// Kitty graphics sparkline (works in Ghostty)
		graph := sparkline.GenerateKittyGraphics(status.WeekValues, status.DailyGoal, 140, 40)
		fmt.Print(graph)
		fmt.Println()
		fmt.Println("       M  T  W  T  F  S  S")

	default:
		ensureDaemon(c, false)
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runDaemon() {
	cfg, _ := config.Load()

	d, err := daemon.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
		os.Exit(1)
	}

	socketPath := storage.SocketPath()
	if err := d.Start(socketPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start socket server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pomme daemon started (socket: %s)\n", socketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if cfg.SketchyBarMode {
		// Headless mode for SketchyBar - no systray menu bar
		<-sigChan
		d.Stop()
		os.Exit(0)
	} else {
		// Traditional mode with systray menu bar
		mb := menubar.New(d)
		go func() {
			<-sigChan
			d.Stop()
			os.Exit(0)
		}()
		mb.Run()
	}
}

func subscript(n int) string {
	if n == 0 {
		return "₀"
	}
	digits := []rune{'₀', '₁', '₂', '₃', '₄', '₅', '₆', '₇', '₈', '₉'}
	var result []rune
	for n > 0 {
		result = append([]rune{digits[n%10]}, result...)
		n /= 10
	}
	return string(result)
}

func ensureDaemon(c *client.Client, silent bool) {
	if c.IsRunning() {
		return
	}

	if !silent {
		fmt.Println("Starting Pomme daemon...")
	}

	cmd := exec.Command(os.Args[0], "--daemon")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Start()

	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if c.IsRunning() {
			if !silent {
				fmt.Println("Daemon started successfully")
			}
			return
		}
	}

	fmt.Fprintln(os.Stderr, "Failed to start daemon")
	os.Exit(1)
}
