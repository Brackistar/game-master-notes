package constants

// Defines the category of a note.
type NoteType uint8

const (
	// Represents a regular free-form note.
	General NoteType = iota
	// Represents a condensed summary note.
	SummaryNote
	// Represents a note that represents a map.
	Map
	// Represents a note focused on a character.
	Character
	// Represents a note focused on a place.
	Location
)

func (n NoteType) String() string {
	return [...]string{"general", "summary_note", "map", "character", "location"}[n]
}

func (n NoteType) EnumIndex() int {
	return int(n)
}
