package constants

type NoteLinkType int8

const (
	Related NoteLinkType = iota
	Contains
	Mentions
	DependsOn
	LocatedIn
)

func (x NoteLinkType) String() string {
	return [...]string{"related", "contains", "mentions", "depends_on", "located_in"}[x]
}
