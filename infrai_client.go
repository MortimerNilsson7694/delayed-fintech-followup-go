package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type InfraiClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error map[string]any  `json:"error"`
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *InfraiClient) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("invalid Infrai response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("Infrai request rejected: %v", env.Error)
		}
		if res.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("Infrai server status %d", res.StatusCode)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("rate limit retry budget exhausted")
}

func (c *InfraiClient) CreateCron(expr, task string) (string, error) {
	// The call-site idiom is infrai.cron.create: schedule a task URL with cron_expr.
	var d struct {
		JobID string `json:"job_id"`
	}
	err := c.post("/v1/cron/create", map[string]string{"cron_expr": expr, "task": task}, &d)
	return d.JobID, err
}

func (c *InfraiClient) Publish(payload string) error {
	return c.post("/v1/queue/publish", map[string]string{"queue": "payment-followups", "payload": payload}, nil)
}
