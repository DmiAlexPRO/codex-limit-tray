package codex

import (
	"codexlimittray/internal/quota"
	"context"
	"fmt"
	"time"
)

type Snapshot struct {
	Windows []quota.Window
	Error   error
	Path    string
	Updated time.Time
	Next    time.Duration
}

func Monitor(ctx context.Context, override string, refresh <-chan struct{}, update func(Snapshot)) {
	failures := 0
	for {
		if ctx.Err() != nil {
			return
		}
		path, err := Resolve(override)
		var windows []quota.Window
		if err == nil {
			windows, err = Read(ctx, path)
		}
		if ctx.Err() != nil {
			return
		}
		failures++
		if err == nil {
			failures = 0
		}
		delay := quota.Delay(failures)
		update(Snapshot{Windows: windows, Error: err, Path: path, Updated: time.Now(), Next: delay})
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-refresh:
			timer.Stop()
		case <-timer.C:
		}
	}
}
func (s Snapshot) Status() string {
	if s.Error != nil {
		return fmt.Sprintf("%s. Повтор через %.0f с", s.Error, s.Next.Seconds())
	}
	return "Обновлено " + s.Updated.Format("15:04:05") + " · запрос каждые 30 с"
}
func (s Snapshot) Diagnostic() string {
	state := "Успешный запрос"
	if s.Error != nil {
		state = s.Error.Error()
	}
	return "Codex: " + s.Path + "\n\n" + state + "\nПоследняя попытка: " + s.Updated.Format("02.01.2006 15:04:05") + "\nЛокальный путь может содержать имя пользователя — скройте его перед публикацией."
}
