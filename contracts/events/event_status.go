package events

type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusExhausted  EventStatus = "exhausted"
	EventStatusPublished  EventStatus = "published"
)
