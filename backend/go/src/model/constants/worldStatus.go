package constants

type WorldStatus uint8

const (
	Draft WorldStatus = iota
	Active
	Archived
)

func (w WorldStatus) String() string {
	return [...]string{"draft", "active", "archived"}[w]
}

func (w WorldStatus) EnumIndex() int {
	return int(w)
}
