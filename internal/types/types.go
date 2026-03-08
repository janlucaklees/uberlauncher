package types

type EntryDTO struct {
	SkillName   string
	EntryID     string
	DisplayText string
	IsFreeText  bool
}

type TriggerType int

const (
	TriggerEntry TriggerType = iota
	TriggerRawInput
)

type RunCommandDTO struct {
	SkillName   string
	EntryID     string
	RawInput    string
	TriggerType TriggerType
}
