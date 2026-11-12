package ui

import (
	"codexlimittray/internal/i18n"
	"codexlimittray/internal/quota"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const MAX_VISIBLE_RESET_ROWS = 5

type resetPanel struct {
	body       *fyne.Container
	button     *widget.Button
	rows       *widget.Label
	scroll     *container.Scroll
	scrollBody *fyne.Container
	credits    *quota.ResetCredits
	redeem     func()
	relayout   func()
	pending    bool
	confirming bool
	message    string
	language   string
}

func newResetPanel(w fyne.Window, language string) *resetPanel {
	p := &resetPanel{body: container.NewVBox(), language: language}
	p.rows = muted("")
	p.scroll = container.NewVScroll(p.rows)
	p.scrollBody = container.New(layout.NewCustomPaddedLayout(0, theme.Size(theme.SizeNameInnerPadding), 0, 0), p.scroll)
	p.button = widget.NewButton(i18n.Text(language, "earnedResetUse"), func() {
		if !p.canRedeem() {
			return
		}
		p.confirming = true
		dialog := dialog.NewCustomConfirm(
			i18n.Text(language, "earnedResetUse"), i18n.Text(language, "earnedResetConfirm"), i18n.Text(language, "earnedResetCancel"),
			wrapped(i18n.Text(language, "earnedResetPrompt")), func(confirmed bool) {
				p.confirming = false
				if !confirmed || !p.canRedeem() {
					return
				}
				p.pending = true
				p.message = "earnedResetPending"
				p.render()
				p.redeem()
			}, w)
		dialog.Show()
	})
	p.body.Hide()
	return p
}

func (v *View) ConfigureReset(redeem func()) {
	v.resets.redeem = redeem
	v.resets.pending = false
	v.resets.message = ""
	v.SetResetCredits(nil, "")
}

func (v *View) SetResetCredits(credits *quota.ResetCredits, message string) {
	p := v.resets
	p.credits = credits
	if message != "" {
		p.pending = false
		p.message = message
	}
	p.render()
}

func (p *resetPanel) canRedeem() bool {
	return !p.pending && !p.confirming && p.redeem != nil && p.credits != nil && p.credits.Available > 0
}

func (p *resetPanel) render() {
	defer p.relayout()
	if p.credits == nil && p.message == "" {
		p.scroll.ScrollToTop()
		p.body.Hide()
		return
	}
	p.body.Objects = nil
	if p.credits != nil {
		if p.canRedeem() {
			p.button.Enable()
		} else {
			p.button.Disable()
		}
		heading := text(i18n.Text(p.language, "earnedResetCount", p.credits.Available), 14, true)
		p.body.Objects = append(p.body.Objects, widget.NewSeparator(), container.NewBorder(nil, nil, nil, p.button, heading))
		rows := make([]string, 0, len(p.credits.ExpiresAt))
		for i, expiry := range p.credits.ExpiresAt {
			row := i18n.Text(p.language, "earnedResetRow", i+1)
			if !expiry.IsZero() {
				row += " · " + i18n.Text(p.language, "earnedResetExpiry", expiry.Local().Format("02.01.2006 · 15:04"))
			}
			rows = append(rows, row)
		}
		p.renderRows(rows)
	}
	if p.message != "" {
		p.body.Objects = append(p.body.Objects, wrapped(i18n.Text(p.language, p.message)))
	}
	p.body.Show()
	p.body.Refresh()
}

func (p *resetPanel) renderRows(rows []string) {
	p.rows.SetText(strings.Join(rows, "\n"))
	if len(rows) > MAX_VISIBLE_RESET_ROWS {
		visible := muted(strings.Join(rows[:MAX_VISIBLE_RESET_ROWS], "\n"))
		// Keep the label's bottom padding outside the clip so the sixth line is fully hidden.
		p.scroll.SetMinSize(fyne.NewSize(0, visible.MinSize().Height-theme.Size(theme.SizeNameInnerPadding)))
		p.body.Objects = append(p.body.Objects, p.scrollBody)
		return
	}
	p.scroll.ScrollToTop()
	if len(rows) > 0 {
		p.body.Objects = append(p.body.Objects, p.rows)
	}
}
