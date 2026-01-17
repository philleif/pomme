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
	startCmd := flag.Bool("start", false, "Start/resume timer")
	pauseCmd := flag.Bool("pause", false, "Pause timer")
	skipCmd := flag.Bool("skip", false, "Skip to next phase")
	resetCmd := flag.Bool("reset", false, "Reset timer")
	toggleBlockCmd := flag.Bool("toggle-block", false, "Toggle Messages blocking")
	statsCmd := flag.Bool("stats", false, "Print today's stats")
	graphCmd := flag.Bool("graph", false, "Show graphical sparkline (Kitty protocol for Ghostty)")

	flag.Parse()

	c := client.New()

	switch {
	case *daemonMode:
		runDaemon()

	case *mcpMode:
		ensureDaemon(c)
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

	case *startCmd:
		ensureDaemon(c)
		_, err := c.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Timer started")

	case *pauseCmd:
		ensureDaemon(c)
		_, err := c.Pause()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Timer paused")

	case *skipCmd:
		ensureDaemon(c)
		_, err := c.Skip()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Skipped to next phase")

	case *resetCmd:
		ensureDaemon(c)
		_, err := c.Reset()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Timer reset")

	case *toggleBlockCmd:
		ensureDaemon(c)
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

	case *statsCmd:
		ensureDaemon(c)
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
		ensureDaemon(c)
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
		ensureDaemon(c)
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runDaemon() {
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

	mb := menubar.New(d)
	go func() {
		<-sigChan
		d.Stop()
		os.Exit(0)
	}()

	mb.Run()
}

func ensureDaemon(c *client.Client) {
	if c.IsRunning() {
		return
	}

	fmt.Println("Starting Pomme daemon...")

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
			fmt.Println("Daemon started successfully")
			return
		}
	}

	fmt.Fprintln(os.Stderr, "Failed to start daemon")
	os.Exit(1)
}
