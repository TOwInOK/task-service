package storage

import (
	"fmt"
	"time"
)

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	TimeSpent   int64      `json:"time_spent,omitempty"` // total seconds spent InProgress
}

// TransitionTo changes the task status while tracking elapsed time.
func (t *Task) TransitionTo(newStatus Status) {
	now := time.Now()

	// Leaving InProgress → accumulate elapsed time
	if t.Status == InProgress && newStatus != InProgress {
		if t.StartedAt != nil {
			t.TimeSpent += int64(now.Sub(*t.StartedAt).Seconds())
			t.StartedAt = nil
		}
	}

	// Entering InProgress → start the clock
	if newStatus == InProgress && t.Status != InProgress {
		t.StartedAt = &now
	}

	t.Status = newStatus
}

// FormatDuration turns seconds into a human-readable string.
func FormatDuration(seconds int64) string {
	if seconds <= 0 {
		return "0s"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
	}
	return fmt.Sprintf("%dd %dh", seconds/86400, (seconds%86400)/3600)
}
