package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/philleif/pomme/internal/daemon"
	"github.com/philleif/pomme/internal/storage"
)

type Client struct {
	socketPath string
}

func New() *Client {
	return &Client{
		socketPath: storage.SocketPath(),
	}
}

func (c *Client) sendCommand(action string) (*daemon.Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("daemon not running (start with 'pomme --daemon')")
	}
	defer conn.Close()

	cmd := daemon.Command{Action: action}
	data, _ := json.Marshal(cmd)
	conn.Write(append(data, '\n'))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp daemon.Response
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &resp, nil
}

func (c *Client) Status() (*daemon.StatusData, error) {
	resp, err := c.sendCommand("status")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status daemon.StatusData
	json.Unmarshal(data, &status)

	return &status, nil
}

func (c *Client) Start() (*daemon.StatusData, error) {
	resp, err := c.sendCommand("start")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status daemon.StatusData
	json.Unmarshal(data, &status)

	return &status, nil
}

func (c *Client) Pause() (*daemon.StatusData, error) {
	resp, err := c.sendCommand("pause")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status daemon.StatusData
	json.Unmarshal(data, &status)

	return &status, nil
}

func (c *Client) Skip() (*daemon.StatusData, error) {
	resp, err := c.sendCommand("skip")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status daemon.StatusData
	json.Unmarshal(data, &status)

	return &status, nil
}

func (c *Client) Reset() (*daemon.StatusData, error) {
	resp, err := c.sendCommand("reset")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status daemon.StatusData
	json.Unmarshal(data, &status)

	return &status, nil
}

func (c *Client) ToggleBlock() (*daemon.StatusData, error) {
	resp, err := c.sendCommand("toggle_block")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status daemon.StatusData
	json.Unmarshal(data, &status)

	return &status, nil
}

func (c *Client) ToggleAlways() (*daemon.StatusData, error) {
	resp, err := c.sendCommand("toggle_always")
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status daemon.StatusData
	json.Unmarshal(data, &status)

	return &status, nil
}

func (c *Client) IsRunning() bool {
	conn, err := net.DialTimeout("unix", c.socketPath, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
