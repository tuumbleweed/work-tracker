package worktracker

import (
	"time"
)

// this is what we save to the JSONL file
type Chunk struct {
	TaskName   string        `json:"task_name"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
	ActiveTime time.Duration `json:"active_time"`
}
