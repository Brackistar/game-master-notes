package constants

// Defines the relationship between two notes.
type NoteLinkType uint8

const (
	// Marks notes as generally related.
	Related NoteLinkType = iota
	// Marks parent-child note containment.
	Contains
	// Marks a reference from one note to another.
	Mentions
	// Marks a dependency relationship.
	DependsOn
	// Marks a spatial or contextual location.
	LocatedIn
)

func (x NoteLinkType) String() string {
	return [...]string{"related", "contains", "mentions", "depends_on", "located_in"}[x]
}
