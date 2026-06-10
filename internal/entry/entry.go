package entry

import (
	"fmt"
	"math"
)

type Key int

const (
	KeyLeft Key = iota
	KeyRight
)

type UI struct {
	KeepOpen      func()
	SetKeyHandler func(func(Key))
}

type Context struct {
	Input string
	UI    UI
}

type Term struct {
	Text     string
	MinScore int
}

type Entry struct {
	Key        string
	SkillId    string
	Label      string
	Terms      []Term
	IsFreeText bool
	Run        func(ctx Context)
}

func (e Entry) Id() string {
	if e.Key != "" {
		return e.Key
	}
	return e.Label
}

func (e Entry) GlobalId() string {
	return fmt.Sprintf("%s:%s", e.SkillId, e.Id())
}

func (e Entry) GetTerms() []Term {
	if len(e.Terms) > 0 {
		return e.Terms
	}
	return []Term{Term{Text: e.Label, MinScore: math.MinInt}}
}
