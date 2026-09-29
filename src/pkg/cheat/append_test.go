package cheat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppend(t *testing.T) {
	for _, newline := range []string{"", "\n"} {
		t.Run("ending="+newline, func(t *testing.T) {
			dir, path, original := fixture(t, newline)
			date, _ := time.Parse("2006-01-02", "2026-09-28")
			chunk, e := Append(dir, date, 15, 60, 75, "")
			if e != nil {
				t.Fatal(e)
			}
			if chunk.TaskName != "Game Development" || chunk.StartedAt.Format(time.RFC3339Nano) != "2026-09-28T17:47:35.165190797-05:00" || chunk.FinishedAt.Sub(chunk.StartedAt) != time.Hour || chunk.ActiveTime != 45*time.Minute {
				t.Fatalf("unexpected chunk: %+v", chunk)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(data), original) {
				t.Fatal("existing data changed")
			}
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			if len(lines) != 2 {
				t.Fatalf("expected two entries: %s", data)
			}
			var saved Chunk
			if err := json.Unmarshal([]byte(lines[1]), &saved); err != nil {
				t.Fatal(err)
			}
			if saved.ActiveTime != chunk.ActiveTime || !saved.StartedAt.Equal(chunk.StartedAt) {
				t.Fatal("incorrect saved entry")
			}
		})
	}
}

func TestInvalidDoesNotModifyFile(t *testing.T) {
	for _, args := range [][3]int64{{-1, 60, 50}, {0, 0, 50}, {0, 60, -1}, {0, 60, 101}, {0, 1000, 50}, {1 << 62, 60, 50}} {
		dir, path, original := fixture(t, "\n")
		date, _ := time.Parse("2006-01-02", "2026-09-28")
		if _, e := Append(dir, date, args[0], args[1], args[2], ""); e == nil {
			t.Fatalf("expected error for %v", args)
		}
		data, _ := os.ReadFile(path)
		if string(data) != original {
			t.Fatal("file modified on invalid input")
		}
	}
}

func TestBadFilesAndTaskOverride(t *testing.T) {
	for _, content := range []string{"", "{bad json}\n", `{"started_at":"2026-09-28T10:00:00Z"}`} {
		dir, path, _ := fixture(t, "\n")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		date, _ := time.Parse("2006-01-02", "2026-09-28")
		if _, e := Append(dir, date, 0, 10, 50, ""); e == nil {
			t.Fatal("expected error")
		}
		data, _ := os.ReadFile(path)
		if string(data) != content {
			t.Fatal("file modified")
		}
	}
	dir, _, _ := fixture(t, "\n")
	date, _ := time.Parse("2006-01-02", "2026-09-28")
	chunk, e := Append(dir, date, 0, 10, 0, "Other")
	if e != nil || chunk.TaskName != "Other" || chunk.ActiveTime != 0 {
		t.Fatalf("override failed: %+v %v", chunk, e)
	}
	if _, e := Append(t.TempDir(), date, 0, 10, 100, ""); e == nil {
		t.Fatal("expected missing-file error")
	}
}

func fixture(t *testing.T, newline string) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "2026", "september", "28_september_2026.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	original := `{"task_name":"Game Development","started_at":"2026-09-28T16:40:05.023050709-05:00","finished_at":"2026-09-28T17:32:35.165190797-05:00","active_time":1862000000123}` + newline
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, path, original
}
