package codex

import (
	"codexlimittray/internal/i18n"
	"codexlimittray/internal/quota"
	"context"
	"errors"
	"time"
)

type Snapshot struct {
	quota.Limits
	Error        error
	Path         string
	Updated      time.Time
	Next         time.Duration
	ResetMessage string
}

func Monitor(ctx context.Context, override string, refresh, redeem <-chan struct{}, update func(Snapshot)) {
	failures := 0
	var resetMessage, resetKey, resetPath string
	for {
		if ctx.Err() != nil {
			return
		}
		path, err := Resolve(override)
		var limits quota.Limits
		if err == nil {
			limits, err = Read(ctx, path)
		}
		if ctx.Err() != nil {
			return
		}
		failures++
		if err == nil {
			failures = 0
		}
		delay := quota.Delay(failures)
		update(Snapshot{Limits: limits, Error: err, Path: path, Updated: time.Now(), Next: delay, ResetMessage: resetMessage})
		resetMessage = ""
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-refresh:
			timer.Stop()
		case <-redeem:
			timer.Stop()
			resetMessage = "earnedResetUnavailable"
			if err != nil || limits.ResetCredits == nil || limits.ResetCredits.Available == 0 {
				continue
			}
			if resetKey == "" || resetPath != path {
				resetKey, resetPath = newResetKey(), path
			}
			resetMessage, err = consumeReset(ctx, path, resetKey)
			if err != nil {
				resetMessage = "earnedResetUnconfirmed"
				continue
			}
			resetKey = ""
		case <-timer.C:
		}
	}
}
func (s Snapshot) Status(language string) string {
	if s.Error != nil {
		var message i18n.MessageError = "codexRead"
		_ = errors.As(s.Error, &message)
		return i18n.Text(language, string(message)) + " · " + i18n.Text(language, "retry", s.Next.Seconds())
	}
	return i18n.Text(language, "updated", s.Updated.Format("15:04:05"))
}

func (s Snapshot) Diagnostic(language string) string {
	state := i18n.Text(language, "requestOK")
	if s.Error != nil {
		state = i18n.ErrorText(language, s.Error)
	}
	return "Codex: " + s.Path + "\n\n" + state + "\n" +
		i18n.Text(language, "lastAttempt", s.Updated.Format("2006-01-02 15:04:05")) + "\n" +
		i18n.Text(language, "privacy")
}
