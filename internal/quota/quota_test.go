package quota

import (
	"testing"
	"time"
)

func TestFormats(t *testing.T) {
	cases := []struct {
		name, payload             string
		count, remaining, minutes int
	}{
		{"single", `{"rateLimits":{"planType":"prolite","primary":{"usedPercent":43,"windowDurationMins":10080}}}`, 1, 57, 10080},
		{"dual", `{"rateLimits":{"primary":{"usedPercent":10,"windowDurationMins":10080},"secondary":{"usedPercent":20,"windowDurationMins":300}}}`, 2, 80, 300},
		{"secondary", `{"rateLimits":{"secondary":{"usedPercent":0}}}`, 1, 100, 0},
		{"clamp", `{"rateLimits":{"primary":{"usedPercent":200,"resetsAt":9223372036854775807}}}`, 1, 0, 0},
		{"authoritative", `{"rateLimitsByLimitId":{"codex":{}},"rateLimits":{"primary":{"usedPercent":1}}}`, 0, 0, 0},
		{"null", `{"rateLimits":{"primary":{"usedPercent":null}}}`, 0, 0, 0},
		{"malformed", `{"rateLimits":`, 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, err := Parse([]byte(c.payload))
			if c.count == 0 {
				if err == nil {
					t.Fatal("missing data accepted")
				}
				return
			}
			if err != nil || len(w) != c.count || w[0].Remaining != c.remaining || w[0].Minutes != c.minutes {
				t.Fatalf("unexpected result: %+v %v", w, err)
			}
		})
	}
}
func TestDelay(t *testing.T) {
	for n, want := range []time.Duration{30, 60, 120, 240, 300, 300} {
		if got := Delay(n); got != want*time.Second {
			t.Fatalf("delay %d: %v", n, got)
		}
	}
}
