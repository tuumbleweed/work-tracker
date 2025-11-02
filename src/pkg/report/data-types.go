package report

import (
	"encoding/json"
	"strings"
	"time"
)



type Chunk struct {
	TaskName   string       `json:"task_name"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	ActiveTime JsonDuration `json:"active_time"`
}

/*
Per-day aggregation used for charts.
*/
type DaySummary struct {
	Date               time.Time                `json:"date"`
	TotalDuration      time.Duration            `json:"total_duration"`
	TotalActive        time.Duration            `json:"total_active"`
	TaskDurations      map[string]time.Duration `json:"task_durations"`
	SmoothedActiveTime time.Duration            `json:"smoothed_active_time"` // Σ (duration * smooth(active_ratio))
}

/*
Top-level aggregation across the whole selected range.
*/
type ReportTotals struct {
	TotalWorked   time.Duration
	TotalActive   time.Duration
	PerTaskTotals map[string]time.Duration
	TaskOrder     []string
}


/*
JSONL input line from work-tracker.

We allow ActiveTime to come either as a JSON number (nanoseconds) or as a string
(parseable by time.ParseDuration, e.g. "999ms", "1.23s").
*/
type JsonDuration struct{ time.Duration }

func (d *JsonDuration) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	// number => nanoseconds
	if len(s) > 0 && (s[0] == '-' || (s[0] >= '0' && s[0] <= '9')) && !strings.ContainsAny(s, `"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`) {
		var n int64
		dec := json.NewDecoder(strings.NewReader(s))
		err := dec.Decode(&n)
		if err != nil {
			return err
		}
		d.Duration = time.Duration(n)
		return nil
	}
	// string => parse duration
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	parsed, perr := time.ParseDuration(str)
	if perr != nil {
		// try to parse bare integer string as nanoseconds
		strTrim := strings.TrimSpace(str)
		if strTrim != "" && strTrim[0] >= '0' && strTrim[0] <= '9' {
			var n2 int64
			dec := json.NewDecoder(strings.NewReader(strTrim))
			e2 := dec.Decode(&n2)
			if e2 == nil {
				d.Duration = time.Duration(n2)
				return nil
			}
		}
		return perr
	}
	d.Duration = parsed
	return nil
}
