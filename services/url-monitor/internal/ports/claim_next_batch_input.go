package ports

import "time"

type ClaimNextBatchInput struct {
	FetchEventsLimit  int
	ProcessingTimeout time.Duration
	Now               time.Time
}
