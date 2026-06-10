package audio

import (
	"fmt"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type Audio struct{}

func New() skill.Skill {
	return &Audio{}
}

func (a *Audio) Id() string { return "audio" }

func (a *Audio) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("wpctl") {
		return
	}

	upsertOutputVolume(ctx)
	upsertInputVolume(ctx)

	sinks, sources, err := listDevices(ctx.Runtime)
	if err != nil {
		ctx.Notifier.ReportError(err)
		return
	}
	for _, s := range sinks {
		s := s
		ctx.Store.UpsertEntry(entry.Entry{
			Key:   "audio:sink:" + s.id,
			Label: "output: " + s.name,
			Run: func(ec entry.Context) {
				if err := ctx.Runtime.Command("wpctl", "set-default", s.id).Run(); err != nil {
					ctx.Notifier.ReportError(err)
				}
			},
		})
	}
	for _, s := range sources {
		s := s
		ctx.Store.UpsertEntry(entry.Entry{
			Key:   "audio:source:" + s.id,
			Label: "input: " + s.name,
			Run: func(ec entry.Context) {
				if err := ctx.Runtime.Command("wpctl", "set-default", s.id).Run(); err != nil {
					ctx.Notifier.ReportError(err)
				}
			},
		})
	}
}

func upsertOutputVolume(ctx skill.Context) {
	ctx.Store.UpsertEntry(entry.Entry{
		Key:   "audio:output-volume",
		Label: volumeLabel("output volume", "@DEFAULT_AUDIO_SINK@", ctx.Runtime),
		Run: func(ec entry.Context) {
			ec.UI.KeepOpen()
			ec.UI.SetKeyHandler(func(key entry.Key) {
				var err error
				switch key {
				case entry.KeyLeft:
					err = ctx.Runtime.Command("wpctl", "set-volume", "@DEFAULT_AUDIO_SINK@", "5%-").Run()
				case entry.KeyRight:
					err = ctx.Runtime.Command("wpctl", "set-volume", "@DEFAULT_AUDIO_SINK@", "5%+").Run()
				default:
					return
				}
				if err != nil {
					ctx.Notifier.ReportError(err)
					return
				}
				upsertOutputVolume(ctx)
			})
		},
	})
}

func upsertInputVolume(ctx skill.Context) {
	ctx.Store.UpsertEntry(entry.Entry{
		Key:   "audio:input-volume",
		Label: volumeLabel("input volume", "@DEFAULT_AUDIO_SOURCE@", ctx.Runtime),
		Run: func(ec entry.Context) {
			ec.UI.KeepOpen()
			ec.UI.SetKeyHandler(func(key entry.Key) {
				var err error
				switch key {
				case entry.KeyLeft:
					err = ctx.Runtime.Command("wpctl", "set-volume", "@DEFAULT_AUDIO_SOURCE@", "5%-").Run()
				case entry.KeyRight:
					err = ctx.Runtime.Command("wpctl", "set-volume", "@DEFAULT_AUDIO_SOURCE@", "5%+").Run()
				default:
					return
				}
				if err != nil {
					ctx.Notifier.ReportError(err)
					return
				}
				upsertInputVolume(ctx)
			})
		},
	})
}

func volumeLabel(prefix, target string, rt skill.Runtime) string {
	out, err := rt.Command("wpctl", "get-volume", target).Output()
	if err != nil {
		return prefix
	}
	// "Volume: 0.75\n" or "Volume: 0.75 [MUTED]\n"
	var vol float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "Volume: %f", &vol); err != nil {
		return prefix
	}
	pct := fmt.Sprintf("%d%%", int(vol*100))
	if strings.Contains(string(out), "[MUTED]") {
		return prefix + " " + pct + " [muted]"
	}
	return prefix + " " + pct
}

type audioDevice struct {
	id   string
	name string
}

func listDevices(rt skill.Runtime) (sinks []audioDevice, sources []audioDevice, err error) {
	out, err := rt.Command("wpctl", "status").Output()
	if err != nil {
		return nil, nil, err
	}

	type section int
	const (
		other section = iota
		sinkSection
		sourceSection
	)

	var current section
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Sinks:") && !strings.Contains(line, "endpoint") {
			current = sinkSection
			continue
		}
		if strings.Contains(line, "Sources:") && !strings.Contains(line, "endpoint") {
			current = sourceSection
			continue
		}
		// A line ending in ":" that isn't a device row signals a new section header.
		// Section headers don't contain the box-drawing continuation character "│".
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ":") && !strings.Contains(line, "│") {
			current = other
			continue
		}
		if current == other {
			continue
		}

		// Strip box-drawing characters so "│   ├─ *52. Name [vol: 1.00]" becomes "52. Name [vol: 1.00]"
		clean := strings.Map(func(r rune) rune {
			switch r {
			case '│', '├', '─', '└', '*':
				return ' '
			}
			return r
		}, line)
		clean = strings.TrimSpace(clean)

		dotIdx := strings.Index(clean, ". ")
		if dotIdx <= 0 {
			continue
		}
		id := strings.TrimSpace(clean[:dotIdx])
		if !isNumeric(id) {
			continue
		}

		rest := clean[dotIdx+2:]
		bracketIdx := strings.LastIndex(rest, "[")
		if bracketIdx <= 0 {
			continue
		}
		name := strings.TrimSpace(rest[:bracketIdx])
		if name == "" {
			continue
		}

		d := audioDevice{id: id, name: name}
		switch current {
		case sinkSection:
			sinks = append(sinks, d)
		case sourceSection:
			sources = append(sources, d)
		}
	}

	return sinks, sources, nil
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
