package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Type string

const (
	Daily       Type = "daily"
	Monthly     Type = "monthly"
	SpecialDate Type = "special_date"
	EvenOdd     Type = "even_odd"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Type        Type      `json:"type"`
	Interval    int64     `json:"interval"`
	ScheduleAt  time.Time `json:"schedule_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (s Type) Valid() bool {
	switch s {
	case Daily, Monthly, SpecialDate, EvenOdd:
		return true
	default:
		return false
	}
}
