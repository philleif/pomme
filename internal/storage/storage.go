package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

type DayStats struct {
	Date      string
	Intervals int
}

func dataDir() (string, error) {
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

func New() (*Storage, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "pomme.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Storage) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS intervals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			completed_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_intervals_date ON intervals(date);
	`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) RecordInterval() error {
	now := time.Now()
	date := now.Format("2006-01-02")
	completedAt := now.Format(time.RFC3339)

	_, err := s.db.Exec(
		"INSERT INTO intervals (date, completed_at) VALUES (?, ?)",
		date, completedAt,
	)
	return err
}

func (s *Storage) TodayCount() (int, error) {
	date := time.Now().Format("2006-01-02")
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM intervals WHERE date = ?",
		date,
	).Scan(&count)
	return count, err
}

func (s *Storage) Last7Days() ([]DayStats, error) {
	stats := make([]DayStats, 7)
	today := time.Now()

	for i := 6; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		var count int
		err := s.db.QueryRow(
			"SELECT COUNT(*) FROM intervals WHERE date = ?",
			dateStr,
		).Scan(&count)
		if err != nil {
			return nil, err
		}

		stats[6-i] = DayStats{
			Date:      dateStr,
			Intervals: count,
		}
	}

	return stats, nil
}

func SocketPath() string {
	dir, err := dataDir()
	if err != nil {
		return "/tmp/pomme.sock"
	}
	return filepath.Join(dir, "pomme.sock")
}
