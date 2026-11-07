package quota

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
)

type Window struct {
	Remaining, Minutes int
	Reset              time.Time
}

func (w Window) Title() string {
	switch w.Minutes {
	case 300:
		return "5 часов"
	case 10080:
		return "Неделя"
	case 0:
		return "Лимит"
	default:
		return fmt.Sprintf("%d мин", w.Minutes)
	}
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

func Parse(data []byte) ([]Window, error) {
	var result struct {
		Legacy  *bucket            `json:"rateLimits"`
		Buckets map[string]*bucket `json:"rateLimitsByLimitId"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.New("неподдерживаемый формат лимитов")
	}
	b := result.Legacy
	if current, present := result.Buckets["codex"]; present {
		b = current
	}
	if b == nil {
		return nil, errors.New("сервис не вернул лимиты Codex")
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
		return nil, errors.New("сервис не вернул распознаваемые окна лимитов")
	}
	sort.SliceStable(windows, func(i, j int) bool {
		return windows[i].Minutes > 0 && windows[j].Minutes > 0 && windows[i].Minutes < windows[j].Minutes
	})
	return windows, nil
}
func Delay(failures int) time.Duration {
	if failures <= 0 {
		return 30 * time.Second
	}
	return time.Duration(min(300, 30<<min(failures, 4))) * time.Second
}
