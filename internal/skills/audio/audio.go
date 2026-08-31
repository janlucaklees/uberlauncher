package audio

import (
	"fmt"
	"regexp"
	"strings"
	"time"

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
	registerStatus(ctx)

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

// registerStatus wires the output/input icon, percentage and active-device
// status bar entries, mirroring how the battery skill reports its own
// icon+percentage pair. dB is only registered when pactl is available, since
// wpctl never exposes it.
func registerStatus(ctx skill.Context) {
	const interval = 1 * time.Second

	ctx.Status.Register("audio:output:icon", interval, func() string {
		return volumeIcon(ctx.Runtime, "@DEFAULT_AUDIO_SINK@", outputIcons)
	})
	ctx.Status.Register("audio:output:percentage", interval, func() string {
		return volumePercentage(ctx.Runtime, "@DEFAULT_AUDIO_SINK@")
	})
	ctx.Status.Register("audio:output:device", interval, func() string {
		return defaultDeviceName(ctx.Runtime, true)
	})

	ctx.Status.Register("audio:input:icon", interval, func() string {
		return volumeIcon(ctx.Runtime, "@DEFAULT_AUDIO_SOURCE@", inputIcons)
	})
	ctx.Status.Register("audio:input:percentage", interval, func() string {
		return volumePercentage(ctx.Runtime, "@DEFAULT_AUDIO_SOURCE@")
	})
	ctx.Status.Register("audio:input:device", interval, func() string {
		return defaultDeviceName(ctx.Runtime, false)
	})

	if ctx.Runtime.HasCommand("pactl") {
		ctx.Status.Register("audio:output:db", interval, func() string {
			db, _ := readDB(ctx.Runtime, "get-sink-volume", "@DEFAULT_SINK@")
			return db
		})
		ctx.Status.Register("audio:input:db", interval, func() string {
			db, _ := readDB(ctx.Runtime, "get-source-volume", "@DEFAULT_SOURCE@")
			return db
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
	pct, muted, ok := readVolume(rt, target)
	if !ok {
		return prefix
	}
	label := fmt.Sprintf("%s %d%%", prefix, pct)
	if muted {
		label += " [muted]"
	}
	return label
}

// readVolume parses wpctl's "Volume: 0.75\n" / "Volume: 0.75 [MUTED]\n" output.
func readVolume(rt skill.Runtime, target string) (pct int, muted bool, ok bool) {
	out, err := rt.Command("wpctl", "get-volume", target).Output()
	if err != nil {
		return 0, false, false
	}
	trimmed := strings.TrimSpace(string(out))
	var vol float64
	if _, err := fmt.Sscanf(trimmed, "Volume: %f", &vol); err != nil {
		return 0, false, false
	}
	return int(vol * 100), strings.Contains(trimmed, "[MUTED]"), true
}

func volumePercentage(rt skill.Runtime, target string) string {
	pct, muted, ok := readVolume(rt, target)
	if !ok {
		return ""
	}
	if muted {
		pct = 0
	}
	return fmt.Sprintf("%d%%", pct)
}

// iconSet holds the muted glyph plus ascending volume-level glyphs for a
// direction (output has loudness tiers, input just has on/off).
type iconSet struct {
	muted string
	tiers []string
}

var outputIcons = iconSet{muted: "󰝟", tiers: []string{"󰕿", "󰖀", "󰕾"}}
var inputIcons = iconSet{muted: "󰍭", tiers: []string{"󰍬"}}

func volumeIcon(rt skill.Runtime, target string, icons iconSet) string {
	pct, muted, ok := readVolume(rt, target)
	if !ok {
		return ""
	}
	if muted {
		return icons.muted
	}
	idx := pct * len(icons.tiers) / 100
	idx = min(max(idx, 0), len(icons.tiers)-1)
	return icons.tiers[idx]
}

var dbPattern = regexp.MustCompile(`-?\d+(\.\d+)?\s*dB`)

// readDB shells out to pactl, which reports dB directly; wpctl never does
// (its volume is a linear 0.0-1.0+ scale over PipeWire's cubic curve, so
// deriving dB from it would not match what other mixers report).
func readDB(rt skill.Runtime, subcommand, target string) (db string, ok bool) {
	out, err := rt.Command("pactl", subcommand, target).Output()
	if err != nil {
		return "", false
	}
	match := dbPattern.FindString(string(out))
	if match == "" {
		return "", false
	}
	return match, true
}

type audioDevice struct {
	id        string
	name      string
	isDefault bool
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

		// "*" marks the current default sink/source, e.g. "│  *   58. Name [vol: 0.58]".
		isDefault := strings.Contains(line, "*")

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

		d := audioDevice{id: id, name: name, isDefault: isDefault}
		switch current {
		case sinkSection:
			sinks = append(sinks, d)
		case sourceSection:
			sources = append(sources, d)
		}
	}

	return sinks, sources, nil
}

// defaultDeviceName returns the name of the currently active sink (wantSinks)
// or source, or "" if none is marked default.
func defaultDeviceName(rt skill.Runtime, wantSinks bool) string {
	sinks, sources, err := listDevices(rt)
	if err != nil {
		return ""
	}
	devices := sources
	if wantSinks {
		devices = sinks
	}
	for _, d := range devices {
		if d.isDefault {
			return d.name
		}
	}
	return ""
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
