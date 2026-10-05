package main

import "testing"

func TestDecideFollowup(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{{"low risk", 12, "notify"}, {"boundary", 70, "review"}, {"high risk", 99, "review"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecideFollowup(PaymentEvent{RiskScore: tt.score}).Action
			if got != tt.want {
				t.Fatalf("score %d: got %s want %s", tt.score, got, tt.want)
			}
		})
	}
}
