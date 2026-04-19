package monitor

import "time"

func newTestMonitor() Monitor {
	return Monitor{
		ID:              1,
		URL:             "https://example1.com",
		IntervalSeconds: 10,
		CreatedAt:       time.Date(2026, 3, 27, 13, 18, 0, 0, time.UTC),
	}
}

func newTestMonitors() []Monitor {
	return []Monitor{
		{
			ID:              1,
			URL:             "https://example1.com",
			IntervalSeconds: 10,
			CreatedAt:       time.Date(2026, 3, 27, 13, 18, 0, 0, time.UTC),
		},
		{
			ID:              2,
			URL:             "https://example2.com",
			IntervalSeconds: 20,
			CreatedAt:       time.Date(2026, 3, 27, 13, 18, 10, 0, time.UTC),
		},
		{
			ID:              3,
			URL:             "https://example3.com",
			IntervalSeconds: 30,
			CreatedAt:       time.Date(2026, 3, 27, 13, 18, 20, 0, time.UTC),
		},
	}
}
