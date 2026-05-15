package health

import (
	"context"
	"net/http"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	OK         bool
	Latency    time.Duration
	Error      string
}

type Checker struct {
	Client *http.Client
}

func (c Checker) Check(ctx context.Context, url string) Result {
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{URL: url, Latency: time.Since(start), Error: err.Error()}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: url, Latency: time.Since(start), Error: err.Error()}
	}
	defer resp.Body.Close()

	return Result{
		URL:        url,
		StatusCode: resp.StatusCode,
		OK:         resp.StatusCode >= 200 && resp.StatusCode < 400,
		Latency:    time.Since(start),
	}
}
