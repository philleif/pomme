package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/philleif/pomme/internal/client"
)

func Run() error {
	s := server.NewMCPServer(
		"Pomme",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	c := client.New()

	statusTool := mcp.NewTool("pomme_status",
		mcp.WithDescription("Get current pomodoro timer status including phase, remaining time, intervals completed today, and weekly sparkline"),
	)
	s.AddTool(statusTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := c.Status()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get status: %v", err)), nil
		}
		data, _ := json.MarshalIndent(status, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})

	startTool := mcp.NewTool("pomme_start",
		mcp.WithDescription("Start or resume the pomodoro timer"),
	)
	s.AddTool(startTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := c.Start()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to start timer: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Timer started. Phase: %s, Remaining: %s", status.Phase, status.Remaining)), nil
	})

	pauseTool := mcp.NewTool("pomme_pause",
		mcp.WithDescription("Pause the pomodoro timer"),
	)
	s.AddTool(pauseTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := c.Pause()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to pause timer: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Timer paused. Phase: %s, Remaining: %s", status.Phase, status.Remaining)), nil
	})

	skipTool := mcp.NewTool("pomme_skip",
		mcp.WithDescription("Skip to the next phase (work -> break or break -> work)"),
	)
	s.AddTool(skipTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := c.Skip()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to skip phase: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Skipped to next phase. Now: %s, Remaining: %s", status.Phase, status.Remaining)), nil
	})

	resetTool := mcp.NewTool("pomme_reset",
		mcp.WithDescription("Reset the pomodoro timer to initial state"),
	)
	s.AddTool(resetTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := c.Reset()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reset timer: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Timer reset. Phase: %s, Remaining: %s", status.Phase, status.Remaining)), nil
	})

	toggleBlockTool := mcp.NewTool("pomme_toggle_block",
		mcp.WithDescription("Toggle Messages.app blocking during focus intervals"),
	)
	s.AddTool(toggleBlockTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := c.ToggleBlock()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to toggle blocking: %v", err)), nil
		}
		state := "OFF"
		if status.BlockEnabled {
			state = "ON"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Messages blocking: %s", state)), nil
	})

	return server.ServeStdio(s)
}
