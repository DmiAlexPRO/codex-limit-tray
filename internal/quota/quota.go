package quota

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"time"
)

type Window struct {
	Remaining, Minutes int
	Reset              time.Time
}

type Limits struct {
	Windows      []Window
	ResetCredits *ResetCredits
}

type ResetCredits struct {
	Available int
	ExpiresAt []time.Time
}

type wireWindow struct {
	Used    *float64 `json:"usedPercent"`
	Minutes int      `json:"windowDurationMins"`
	Reset   *int64   `json:"resetsAt"`
}
type bucket struct {
	Primary   *wireWindow `json:"primary"`
	Secondary *wireWindow `json:"secondary"`
}

func Parse(data []byte) (Limits, error) {
	var result struct {
		Legacy  *bucket            `json:"rateLimits"`
		Buckets map[string]*bucket `json:"rateLimitsByLimitId"`
		Credits json.RawMessage    `json:"rateLimitResetCredits"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return Limits{}, errors.New("unsupported quota format")
	}
	b := result.Legacy
	if current, present := result.Buckets["codex"]; present {
		b = current
	}
	if b == nil {
		return Limits{}, errors.New("Codex returned no quota data")
	}
	windows := make([]Window, 0, 2)
	for _, source := range []*wireWindow{b.Primary, b.Secondary} {
		if source == nil || source.Used == nil || math.IsNaN(*source.Used) || math.IsInf(*source.Used, 0) {
			continue
		}
		w := Window{Remaining: int(math.Round(math.Max(0, math.Min(100, 100-*source.Used)))), Minutes: max(0, source.Minutes)}
		if source.Reset != nil && *source.Reset >= 0 && *source.Reset <= 253402214399 {
			w.Reset = time.Unix(*source.Reset, 0)
		}
		windows = append(windows, w)
	}
	if len(windows) == 0 {
		return Limits{}, errors.New("Codex returned no recognized quota windows")
	}
	sort.SliceStable(windows, func(i, j int) bool {
		return windows[i].Minutes > 0 && windows[j].Minutes > 0 && windows[i].Minutes < windows[j].Minutes
	})
	return Limits{Windows: windows, ResetCredits: parseResetCredits(result.Credits)}, nil
}

func parseResetCredits(data json.RawMessage) *ResetCredits {
	var wire struct {
		Available *int `json:"availableCount"`
		Credits   []struct {
			Status    string `json:"status"`
			ExpiresAt *int64 `json:"expiresAt"`
		} `json:"credits"`
	}
	if json.Unmarshal(data, &wire) != nil || wire.Available == nil || *wire.Available < 0 {
		return nil
	}
	credits := &ResetCredits{Available: *wire.Available}
	for _, credit := range wire.Credits {
		if len(credits.ExpiresAt) >= credits.Available {
			break
		}
		if credit.Status != "available" {
			continue
		}
		var expiry time.Time
		if credit.ExpiresAt != nil && *credit.ExpiresAt >= 0 && *credit.ExpiresAt <= 253402214399 {
			expiry = time.Unix(*credit.ExpiresAt, 0)
		}
		credits.ExpiresAt = append(credits.ExpiresAt, expiry)
	}
	return credits
}
func Delay(failures int) time.Duration {
	if failures <= 0 {
		return 30 * time.Second
	}
	return time.Duration(min(300, 30<<min(failures, 4))) * time.Second
}
