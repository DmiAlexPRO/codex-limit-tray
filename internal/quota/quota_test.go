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
			limits, err := Parse([]byte(c.payload))
			w := limits.Windows
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

func TestResetCredits(t *testing.T) {
	for _, tc := range []struct {
		credits string
		count   int
		rows    int
	}{
		{"null", -1, 0},
		{`{"credits":[]}`, -1, 0},
		{`{"availableCount":"bad"}`, -1, 0},
		{`{"availableCount":0,"credits":[]}`, 0, 0},
		{`{"availableCount":4,"credits":null}`, 4, 0},
		{`{"availableCount":3,"credits":[{"status":"redeemed"},{"status":"available","expiresAt":1794484800},{"status":"available","expiresAt":null}]}`, 3, 2},
	} {
		t.Run(tc.credits, func(t *testing.T) {
			limits, err := Parse([]byte(`{"rateLimits":{"primary":{"usedPercent":43}},"rateLimitResetCredits":` + tc.credits + `}`))
			if err != nil || len(limits.Windows) != 1 {
				t.Fatalf("lost quota windows: %+v %v", limits, err)
			}
			credits := limits.ResetCredits
			if tc.count < 0 {
				if credits != nil {
					t.Fatalf("expected unknown count, got %+v", credits)
				}
				return
			}
			if credits == nil || credits.Available != tc.count || len(credits.ExpiresAt) != tc.rows {
				t.Fatalf("unexpected credits: %+v", credits)
			}
			if tc.rows == 2 && (credits.ExpiresAt[0].Unix() != 1794484800 || !credits.ExpiresAt[1].IsZero()) {
				t.Fatalf("unexpected expiry: %v", credits.ExpiresAt)
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
