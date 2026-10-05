package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PaymentEvent struct {
	ID          string `json:"id"`
	Customer    string `json:"customer"`
	AmountCents int    `json:"amount_cents"`
	RiskScore   int    `json:"risk_score"`
}
type FollowupDecision struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

func DecideFollowup(e PaymentEvent) FollowupDecision {
	if e.RiskScore >= 70 {
		return FollowupDecision{"review", "risk score requires manual review"}
	}
	return FollowupDecision{"notify", "payment follow-up is eligible"}
}

func scheduleFollowup(c *InfraiClient, e PaymentEvent, hours int, webhook string) (FollowupDecision, error) {
	d := DecideFollowup(e)
	payload, _ := json.Marshal(map[string]any{"event": e, "decision": d, "scheduled_at": time.Now().UTC()})
	if err := c.Publish(string(payload)); err != nil {
		return d, err
	}
	if d.Action == "notify" {
		_, err := c.CreateCron(fmt.Sprintf("0 */%d * * *", hours), webhook)
		if err != nil {
			return d, err
		}
	}
	return d, nil
}

func followupHandler(c *InfraiClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var e PaymentEvent
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		d, err := scheduleFollowup(c, e, 4, "https://example.com/fintech/followup")
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	}
}
