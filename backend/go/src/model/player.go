package model

// Represents a participant in one or more campaigns.
type Player struct {
	ID          ULID
	Name        string
	AuditFields AuditFields
}
