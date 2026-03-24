package monitor

import (
	"context"
	"net/http"
	"time"
)

type Checker struct {}

func (c *Checker) Check(ctx context.Context, m Monitor) MonitorCheck {
	client := &http.Client{
		Timeout: time.Duration(m.IntervalSeconds/2) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	check := MonitorCheck{
		MonitorID: m.ID,
		StartedAt: time.Now(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)
	if err != nil {
		check.Status = MonitorCheckStatusError
		check.ErrorMessage = err.Error()
		check.FinishedAt = time.Now()
		check.ResponseTimeMS = check.FinishedAt.Sub(check.StartedAt).Milliseconds()
		return check
	}

	resp, err := client.Do(req)
	if err != nil {
		check.Status = MonitorCheckStatusError
		check.ErrorMessage = err.Error()
		check.FinishedAt = time.Now()
		check.ResponseTimeMS = check.FinishedAt.Sub(check.StartedAt).Milliseconds()
		return check
	}
	defer resp.Body.Close()

	check.FinishedAt = time.Now()
	check.ResponseTimeMS = check.FinishedAt.Sub(check.StartedAt).Milliseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		check.Status = MonitorCheckStatusUp
	} else {
		check.Status = MonitorCheckStatusDown
	}

	check.HTTPStatusCode = int16(resp.StatusCode)

	return check
}
