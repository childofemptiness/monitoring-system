package monitor

import "errors"

var (
	ErrInvalidURL           = errors.New("invalid url")
	ErrInvalidInterval      = errors.New("invalid interval")
	ErrMonitorAlreadyExists = errors.New("monitor already exists")
	ErrMonitorNotFound      = errors.New("monitor not found")
)
