package constants

// Defines the lifecycle state of a world.
type WorldStatus uint8

const (
	// Represents a world still in preparation.
	Draft WorldStatus = iota
	// Represents a world currently in use.
	Active
	// Represents a world kept for history.
	Archived
)

func (w WorldStatus) String() string {
	return [...]string{"draft", "active", "archived"}[w]
}

func (w WorldStatus) EnumIndex() int {
	return int(w)
}
