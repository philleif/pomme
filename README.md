# Pomme 🍅

A macOS pomodoro timer with Messages app blocking, menu bar integration, and TUI controls. Designed for seamless tmux, SketchyBar, simple-bar, and MCP integration.

## Features

- **Pomodoro Timer**: 30-min work / 5-min break / 20-min long break (configurable)
- **App Blocking**: Automatically blocks Messages.app during focus intervals
- **Menu Bar**: Live timer display with weekly sparkline via macOS system tray
- **TUI Interface**: Compact terminal UI built with Bubble Tea
- **tmux Integration**: Status line output and keybinding commands
- **SketchyBar Integration**: JSON output for SketchyBar widgets with headless daemon mode
- **simple-bar Integration**: Compact status output for Ubersicht simple-bar widgets
- **MCP Server**: AI assistant integration via Model Context Protocol (stdio)
- **macOS App Bundle**: Standalone `.app` with Kitty/Terminal.app launcher
- **Tufte-inspired Sparklines**: Braille characters or Kitty graphics (for Ghostty)
- **Notifications**: macOS notifications on phase transitions

## Prerequisites

- **macOS** 11.0+ (required for Messages blocking, systray, notifications)
- **Go** 1.23+
- **CGO enabled** (required by `go-sqlite3`; enabled by default on macOS)

## Installation

### CLI Binary

```bash
make build      # Build binary to ./build/pomme
make install    # Build and copy to /usr/local/bin/pomme
```

### macOS App Bundle

```bash
make app            # Build Pomme.app in ./macos/
make app-install    # Copy Pomme.app to /Applications/
```

The app bundle includes a launcher script that prefers Kitty (with native float-on-top support) and falls back to Terminal.app.

### Other Make Targets

```bash
make deps       # Tidy and download Go dependencies
make test       # Run tests
make clean      # Remove build artifacts and socket file
make run        # Build and run TUI
make daemon     # Build and run daemon directly
```

## Usage

### Quick Start

```bash
pomme             # Opens TUI and auto-starts daemon in background
```

The daemon provides the menu bar icon and runs persistently. It auto-starts when you run any pomme command.

### TUI Controls

| Key | Action                       |
|-----|------------------------------|
| `s` | Start timer                  |
| `p` | Pause timer                  |
| `k` | Skip to next phase           |
| `r` | Reset timer                  |
| `b` | Toggle Messages blocking     |
| `a` | Toggle "always block" mode   |
| `q` | Quit TUI                     |

### Command Line

```bash
pomme --status          # Print status line (for tmux)
pomme --start           # Start/resume timer
pomme --pause           # Pause timer
pomme --skip            # Skip to next phase
pomme --reset           # Reset timer
pomme --toggle          # Toggle start/pause
pomme --toggle-block    # Toggle Messages blocking
pomme --toggle-always   # Toggle "always block Messages" mode
pomme --stats           # Print today's stats with braille sparkline
pomme --graph           # Show pixel-based sparkline (Kitty graphics for Ghostty)
pomme --daemon          # Run as daemon (menu bar only, no TUI)
pomme --simplebar       # Print compact status for simple-bar widget
pomme --sketchybar      # Print JSON status for SketchyBar widget
pomme --mcp             # Run as MCP server (stdio)
```

## macOS Menu Bar

By default, the Pomme daemon displays a live timer in the macOS menu bar using the system tray. This works out of the box -- no SketchyBar or other third-party status bar required.

The menu bar icon shows the current timer status (e.g. `🍅 18:32 ▃▅▇▆▄▂█`) and provides a dropdown menu with controls:

- **Start / Pause / Skip / Reset** the timer
- **Block Messages** toggle
- **Always Block** toggle
- **Open Controls** -- opens the TUI in Terminal.app

The daemon auto-starts when you run any `pomme` command. To run the daemon standalone (menu bar only, no TUI):

```bash
pomme --daemon
```

If you use SketchyBar and want to disable the built-in menu bar, set `"sketchybar_mode": true` in your config -- the daemon will then run headless.

## tmux Integration

### Status Bar

Add to your `~/.tmux.conf`:

```bash
set -g status-right "#(pomme --status) | %H:%M"
set -g status-interval 5
```

Output example: `🍅 18:32 ▃▅▇▆▄▂█`

### Keybindings

```bash
bind-key P run-shell "pomme --start"
bind-key O run-shell "pomme --pause"
bind-key K run-shell "pomme --skip"
```

## SketchyBar Integration

Use `--sketchybar` for JSON output suitable for SketchyBar items:

```bash
pomme --sketchybar
```

Example output:

```json
{"icon":"󰔟","label":"18:32 ₃","color":"0xff9ece6a","block_icon":"●","block_color":"0xff9ece6a","drawing":"on"}
```

To run the daemon without the macOS systray menu bar (headless mode for SketchyBar), set `"sketchybar_mode": true` in your config. The daemon will run headless and rely entirely on SketchyBar for display.

## simple-bar Integration

Use `--simplebar` for compact output suitable for Ubersicht simple-bar widgets:

```bash
pomme --simplebar
```

Output format: `🍅 18:32 ₃`

Related config fields: `simplebar_enabled`, `simplebar_widget_id`, `simplebar_port`.

## skhd Integration

Bind global hotkeys to Pomme commands using [skhd](https://github.com/koekeishiya/skhd). Add to your `~/.skhdrc`:

```bash
ctrl + alt - s : pomme --toggle
ctrl + alt - k : pomme --skip
ctrl + alt - r : pomme --reset
ctrl + alt - b : pomme --toggle-block
ctrl + alt - a : pomme --toggle-always
```

The daemon auto-starts on first command, so no setup beyond the keybindings is needed.

## MCP Integration

Pomme includes an MCP (Model Context Protocol) server for AI assistant integration. Run it via stdio:

```bash
pomme --mcp
```

### Claude Desktop Configuration

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "pomme": {
      "command": "pomme",
      "args": ["--mcp"]
    }
  }
}
```

### Available MCP Tools

| Tool                  | Description                                      |
|-----------------------|--------------------------------------------------|
| `pomme_status`        | Get current timer status, phase, remaining time, intervals, and weekly sparkline |
| `pomme_start`         | Start or resume the timer                        |
| `pomme_pause`         | Pause the timer                                  |
| `pomme_skip`          | Skip to the next phase                           |
| `pomme_reset`         | Reset the timer                                  |
| `pomme_toggle_block`  | Toggle Messages.app blocking                     |

## Configuration

Pomme creates a config file at `~/.pomme/config.json` on first run:

```json
{
  "work_duration_minutes": 30,
  "short_break_duration_minutes": 5,
  "long_break_duration_minutes": 20,
  "long_break_after_intervals": 4,
  "daily_goal": 12,
  "block_messages_enabled": true,
  "always_block": false,
  "simplebar_enabled": false,
  "simplebar_widget_id": 1,
  "simplebar_port": 7776,
  "sketchybar_mode": false,
  "notifications_enabled": true
}
```

Edit this file to customize your intervals and integrations. Restart the daemon for changes to take effect.

## Data Storage

- Config: `~/.pomme/config.json`
- Database: `~/.pomme/pomme.db`
- Socket: `~/.pomme/pomme.sock`

## Sparkline Display

Pomme uses **Tufte-inspired sparklines** to show your weekly progress. Day labels are dynamic based on today's day of week.

### Braille Sparkline (default)
Uses Unicode braille characters for smooth 8-level resolution:
```
Week:  ⣀  ⣤  ⣶  ⣿  ⣷  ⣄  ⣿
       S  M  T  W  T  F  S
       2  5  9  12 10 4  8
                        ↑today
```

### Kitty Graphics (for Ghostty)
Use `--graph` flag for pixel-based sparklines using the Kitty graphics protocol:
```bash
pomme --graph
```
This renders actual pixel graphics in terminals that support it (Ghostty, Kitty).

### Status Line (for tmux)
Compact format with subscript digits:
```
🍅 18:32 ⣀ ⣤ ⣶ ⣿ ⣷ ⣄ ⣿₈
```

## Pomodoro Best Practices

Based on research:

- **30-minute work intervals** (default) - balance focus and sustainability
- **5-minute breaks** allow mental recovery
- **Long break (20 min) after 4 intervals** prevents fatigue
- **12 intervals/day** = ~6 hours of deep focused work
- **Block distractions** - Messages blocking prevents context switching

## License

MIT
