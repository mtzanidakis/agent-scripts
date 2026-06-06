package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func dbPath() string {
	if p := os.Getenv("GH_DAILY_DB"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gh-daily-tasks.db"
	}
	return filepath.Join(home, ".gh-daily-tasks.db")
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath())
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS repo_snapshots (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			full_name   TEXT    NOT NULL,
			stars       INTEGER NOT NULL,
			forks       INTEGER NOT NULL,
			recorded_at TEXT    NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_repo_snapshots_name_time
			ON repo_snapshots(full_name, recorded_at DESC);
	`)
	return err
}

type Snapshot struct {
	FullName string
	Stars    int
	Forks    int
}

func getPreviousSnapshots(db *sql.DB, repoNames []string) (map[string]Snapshot, error) {
	result := make(map[string]Snapshot)
	if len(repoNames) == 0 {
		return result, nil
	}

	// Query last snapshot for each repo individually (simpler than dynamic IN clause)
	stmt, err := db.Prepare(`
		SELECT stars, forks FROM repo_snapshots
		WHERE full_name = ?
		ORDER BY recorded_at DESC
		LIMIT 1
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = stmt.Close() }()

	for _, name := range repoNames {
		var s Snapshot
		s.FullName = name
		err := stmt.QueryRow(name).Scan(&s.Stars, &s.Forks)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("querying snapshot for %s: %w", name, err)
		}
		result[name] = s
	}
	return result, nil
}

func saveSnapshots(db *sql.DB, reports []RepoReport) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`INSERT INTO repo_snapshots (full_name, stars, forks, recorded_at) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, r := range reports {
		if _, err := stmt.Exec(r.FullName, r.Stars, r.Forks, now); err != nil {
			return fmt.Errorf("inserting snapshot for %s: %w", r.FullName, err)
		}
	}
	return tx.Commit()
}

// computeDeltas returns star/fork deltas for each repo compared to previous snapshot.
func computeDeltas(reports []RepoReport, previous map[string]Snapshot) map[string][2]int {
	deltas := make(map[string][2]int) // [starsDelta, forksDelta]
	for _, r := range reports {
		prev, ok := previous[r.FullName]
		if !ok {
			deltas[r.FullName] = [2]int{0, 0}
			continue
		}
		deltas[r.FullName] = [2]int{r.Stars - prev.Stars, r.Forks - prev.Forks}
	}
	return deltas
}
