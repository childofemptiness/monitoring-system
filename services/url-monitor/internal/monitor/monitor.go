package monitor

import "time"

type Monitor struct {
	ID              int64
	URL             string
	IntervalSeconds int
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	LastCheckAt     *time.Time
	NextCheckAt     *time.Time
}
