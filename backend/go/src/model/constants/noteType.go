package constants

type NoteType int8

const (
	General NoteType = iota
	SummaryNote
	Map
	Character
	Location
)

func (n NoteType) String() string {
	return [...]string{"general", "summary_note", "map", "character", "location"}[n]
}

func (n NoteType) EnumIndex() int {
	return int(n)
}
