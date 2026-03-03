package main

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrate(t *testing.T) {
	db := setupTestDB(t)

	// Verify table exists by inserting a row
	_, err := db.Exec(`INSERT INTO repo_snapshots (full_name, stars, forks, recorded_at) VALUES ('a/b', 1, 2, '2024-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert after migrate: %v", err)
	}

	// Running migrate again should be idempotent
	if err := migrate(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

func TestSaveAndGetSnapshots(t *testing.T) {
	db := setupTestDB(t)

	reports := []RepoReport{
		{FullName: "user/repo-a", Stars: 10, Forks: 3},
		{FullName: "user/repo-b", Stars: 20, Forks: 5},
	}

	if err := saveSnapshots(db, reports); err != nil {
		t.Fatalf("saveSnapshots: %v", err)
	}

	prev, err := getPreviousSnapshots(db, []string{"user/repo-a", "user/repo-b", "user/nonexistent"})
	if err != nil {
		t.Fatalf("getPreviousSnapshots: %v", err)
	}

	if len(prev) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(prev))
	}
	if s := prev["user/repo-a"]; s.Stars != 10 || s.Forks != 3 {
		t.Errorf("repo-a: got stars=%d forks=%d, want 10/3", s.Stars, s.Forks)
	}
	if s := prev["user/repo-b"]; s.Stars != 20 || s.Forks != 5 {
		t.Errorf("repo-b: got stars=%d forks=%d, want 20/5", s.Stars, s.Forks)
	}
	if _, ok := prev["user/nonexistent"]; ok {
		t.Error("nonexistent repo should not be in snapshots")
	}
}

func TestGetPreviousSnapshotsReturnsLatest(t *testing.T) {
	db := setupTestDB(t)

	// Insert two snapshots for the same repo — getPreviousSnapshots should return the latest
	_, err := db.Exec(`INSERT INTO repo_snapshots (full_name, stars, forks, recorded_at) VALUES
		('user/repo', 5, 1, '2024-01-01T00:00:00Z'),
		('user/repo', 10, 2, '2024-01-02T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}

	prev, err := getPreviousSnapshots(db, []string{"user/repo"})
	if err != nil {
		t.Fatal(err)
	}
	if s := prev["user/repo"]; s.Stars != 10 || s.Forks != 2 {
		t.Errorf("got stars=%d forks=%d, want 10/2 (latest snapshot)", s.Stars, s.Forks)
	}
}

func TestGetPreviousSnapshotsEmpty(t *testing.T) {
	db := setupTestDB(t)

	prev, err := getPreviousSnapshots(db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(prev) != 0 {
		t.Errorf("expected empty map, got %d entries", len(prev))
	}
}

func TestComputeDeltas(t *testing.T) {
	reports := []RepoReport{
		{FullName: "user/grew", Stars: 15, Forks: 4},
		{FullName: "user/shrank", Stars: 8, Forks: 2},
		{FullName: "user/new", Stars: 1, Forks: 0},
	}
	previous := map[string]Snapshot{
		"user/grew":   {Stars: 10, Forks: 3},
		"user/shrank": {Stars: 12, Forks: 5},
		// "user/new" has no previous snapshot
	}

	deltas := computeDeltas(reports, previous)

	tests := []struct {
		name                 string
		wantStars, wantForks int
	}{
		{"user/grew", 5, 1},
		{"user/shrank", -4, -3},
		{"user/new", 0, 0},
	}
	for _, tt := range tests {
		d := deltas[tt.name]
		if d[0] != tt.wantStars || d[1] != tt.wantForks {
			t.Errorf("%s: got delta [%d, %d], want [%d, %d]", tt.name, d[0], d[1], tt.wantStars, tt.wantForks)
		}
	}
}

func TestDbPath(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		t.Setenv("GH_DAILY_DB", "/tmp/custom.db")
		if got := dbPath(); got != "/tmp/custom.db" {
			t.Errorf("got %q, want /tmp/custom.db", got)
		}
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("GH_DAILY_DB", "")
		home, _ := os.UserHomeDir()
		got := dbPath()
		want := home + "/.gh-daily-tasks.db"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
