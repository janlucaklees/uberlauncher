package types

type Entry struct {
	SkillName        string
	EntryID          string
	DisplayText      string
	SupportsFreeText bool
}

func NewEntry(skillName, entryID string) Entry {
	return Entry{
		SkillName:        skillName,
		EntryID:          entryID,
		DisplayText:      skillName + " " + entryID,
		SupportsFreeText: false,
	}
}

type Command struct {
	Entry    Entry
	RawInput string
}
