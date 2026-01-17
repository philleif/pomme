package blocker

import (
	"os/exec"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	messagesProcessName = "Messages"
	checkInterval       = 1 * time.Second
)

type Blocker struct {
	mu          sync.RWMutex
	enabled     bool
	alwaysBlock bool
	inInterval  bool
	stopChan    chan struct{}
	running     bool
}

func New() *Blocker {
	return &Blocker{}
}

func (b *Blocker) Start() {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	b.running = true
	b.stopChan = make(chan struct{})
	b.mu.Unlock()

	go b.run()
}

func (b *Blocker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running && b.stopChan != nil {
		close(b.stopChan)
		b.running = false
	}
}

func (b *Blocker) run() {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopChan:
			return
		case <-ticker.C:
			if b.shouldBlock() {
				b.killMessages()
			}
		}
	}
}

func (b *Blocker) shouldBlock() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.enabled {
		return false
	}

	return b.inInterval || b.alwaysBlock
}

func (b *Blocker) killMessages() {
	processes, err := process.Processes()
	if err != nil {
		return
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		if name == messagesProcessName {
			p.Kill()
			b.sendNotification()
		}
	}
}

func (b *Blocker) sendNotification() {
	script := `display notification "Messages is blocked during focus time" with title "Pomme"`
	exec.Command("osascript", "-e", script).Run()
}

func (b *Blocker) SetEnabled(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = enabled
}

func (b *Blocker) Enabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

func (b *Blocker) SetAlwaysBlock(always bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.alwaysBlock = always
}

func (b *Blocker) AlwaysBlock() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.alwaysBlock
}

func (b *Blocker) SetInInterval(inInterval bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inInterval = inInterval
}

func (b *Blocker) ToggleEnabled() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = !b.enabled
	return b.enabled
}

func (b *Blocker) ToggleAlwaysBlock() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.alwaysBlock = !b.alwaysBlock
	return b.alwaysBlock
}
