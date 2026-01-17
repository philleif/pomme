package timer

import (
	"sync"
	"time"
)

type Phase int

const (
	PhaseWork Phase = iota
	PhaseShortBreak
	PhaseLongBreak
)

func (p Phase) String() string {
	switch p {
	case PhaseWork:
		return "work"
	case PhaseShortBreak:
		return "short_break"
	case PhaseLongBreak:
		return "long_break"
	default:
		return "unknown"
	}
}

type State int

const (
	StateIdle State = iota
	StateRunning
	StatePaused
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	default:
		return "unknown"
	}
}

type Config struct {
	WorkDuration       time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration  time.Duration
	LongBreakAfter     int // Number of work intervals before long break
}

func DefaultConfig() Config {
	return Config{
		WorkDuration:       30 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  20 * time.Minute,
		LongBreakAfter:     4,
	}
}

type Timer struct {
	mu     sync.RWMutex
	config Config

	state           State
	phase           Phase
	remaining       time.Duration
	intervalsToday  int
	intervalsSinceBreak int

	lastTick   time.Time
	onComplete func(phase Phase)
	stopChan   chan struct{}
}

func New(config Config) *Timer {
	return &Timer{
		config:    config,
		state:     StateIdle,
		phase:     PhaseWork,
		remaining: config.WorkDuration,
	}
}

func (t *Timer) SetOnComplete(fn func(phase Phase)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onComplete = fn
}

func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateRunning {
		return
	}

	t.state = StateRunning
	t.lastTick = time.Now()
	t.stopChan = make(chan struct{})

	go t.run()
}

func (t *Timer) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopChan:
			return
		case now := <-ticker.C:
			t.mu.Lock()
			if t.state != StateRunning {
				t.mu.Unlock()
				continue
			}

			elapsed := now.Sub(t.lastTick)
			t.lastTick = now
			t.remaining -= elapsed

			if t.remaining <= 0 {
				completedPhase := t.phase
				t.advancePhase()
				onComplete := t.onComplete
				t.mu.Unlock()

				if onComplete != nil {
					onComplete(completedPhase)
				}
			} else {
				t.mu.Unlock()
			}
		}
	}
}

func (t *Timer) advancePhase() {
	if t.phase == PhaseWork {
		t.intervalsToday++
		t.intervalsSinceBreak++

		if t.intervalsSinceBreak >= t.config.LongBreakAfter {
			t.phase = PhaseLongBreak
			t.remaining = t.config.LongBreakDuration
			t.intervalsSinceBreak = 0
		} else {
			t.phase = PhaseShortBreak
			t.remaining = t.config.ShortBreakDuration
		}
	} else {
		t.phase = PhaseWork
		t.remaining = t.config.WorkDuration
	}
}

func (t *Timer) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateRunning {
		t.state = StatePaused
		if t.stopChan != nil {
			close(t.stopChan)
		}
	}
}

func (t *Timer) Resume() {
	t.mu.Lock()
	if t.state != StatePaused {
		t.mu.Unlock()
		return
	}
	t.mu.Unlock()
	t.Start()
}

func (t *Timer) Skip() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateRunning && t.stopChan != nil {
		close(t.stopChan)
	}

	completedPhase := t.phase
	t.advancePhase()
	t.state = StateIdle

	if t.onComplete != nil {
		go t.onComplete(completedPhase)
	}
}

func (t *Timer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateRunning && t.stopChan != nil {
		close(t.stopChan)
	}

	t.state = StateIdle
	t.phase = PhaseWork
	t.remaining = t.config.WorkDuration
	t.intervalsSinceBreak = 0
}

func (t *Timer) State() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

func (t *Timer) Phase() Phase {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.phase
}

func (t *Timer) Remaining() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.remaining
}

func (t *Timer) IntervalsToday() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.intervalsToday
}

func (t *Timer) SetIntervalsToday(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.intervalsToday = n
}

type Status struct {
	State          State
	Phase          Phase
	Remaining      time.Duration
	IntervalsToday int
}

func (t *Timer) Status() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return Status{
		State:          t.state,
		Phase:          t.phase,
		Remaining:      t.remaining,
		IntervalsToday: t.intervalsToday,
	}
}
