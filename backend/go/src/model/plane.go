package model

// Represents a distinct reality layer within a world.
type Plane struct {
	ID          ULID
	Name        string
	Description string
	AuditFields AuditFields
}
