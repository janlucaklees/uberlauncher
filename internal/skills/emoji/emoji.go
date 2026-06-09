package emoji

import (
	_ "embed"
	"errors"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

//go:embed emojis.txt
var emojiData string

const minScore = 5

type Emoji struct{}

func New() skill.Skill {
	return &Emoji{}
}

func (e *Emoji) Id() string {
	return "emoji"
}

func (e *Emoji) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("wl-copy") {
		ctx.Notifier.ReportError(errors.New("wl-copy not found"))
		return
	}

	for _, line := range strings.Split(emojiData, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		char, name := parts[0], parts[1]
		label := ":" + name + ":"
		ctx.Store.UpsertEntry(entry.Entry{
			Label: char,
			Terms: []entry.Term{{Text: label, MinScore: minScore}},
			Run: func(ec entry.Context) {
				if err := ctx.Runtime.Command("wl-copy", char).Run(); err != nil {
					ec.UI.KeepOpen()
					ctx.Notifier.ReportError(err)
				}
			},
		})
	}
}
