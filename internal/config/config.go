package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	WorkDuration       int  `json:"work_duration_minutes"`
	ShortBreakDuration int  `json:"short_break_duration_minutes"`
	LongBreakDuration  int  `json:"long_break_duration_minutes"`
	LongBreakAfter     int  `json:"long_break_after_intervals"`
	DailyGoal          int  `json:"daily_goal"`
	BlockMessages      bool `json:"block_messages_enabled"`
	AlwaysBlock        bool `json:"always_block"`
	SimpleBarEnabled   bool `json:"simplebar_enabled"`
	SimpleBarWidgetID  int  `json:"simplebar_widget_id"`
	SimpleBarPort      int  `json:"simplebar_port"`
}

func Default() Config {
	return Config{
		WorkDuration:       30,
		ShortBreakDuration: 5,
		LongBreakDuration:  20,
		LongBreakAfter:     4,
		DailyGoal:          12,
		BlockMessages:      true,
		AlwaysBlock:        false,
		SimpleBarEnabled:   false,
		SimpleBarWidgetID:  1,
		SimpleBarPort:      7776,
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".pomme")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func ConfigPath() string {
	dir, err := configDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "config.json")
}

func Load() (Config, error) {
	cfg := Default()

	path := ConfigPath()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config file
			Save(cfg)
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}

	return cfg, nil
}

func Save(cfg Config) error {
	path := ConfigPath()
	if path == "" {
		return nil
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c Config) WorkDurationTime() time.Duration {
	return time.Duration(c.WorkDuration) * time.Minute
}

func (c Config) ShortBreakDurationTime() time.Duration {
	return time.Duration(c.ShortBreakDuration) * time.Minute
}

func (c Config) LongBreakDurationTime() time.Duration {
	return time.Duration(c.LongBreakDuration) * time.Minute
}
