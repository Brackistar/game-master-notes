package model

// Represents a distinct reality layer within a world.
type Plane struct {
	ID          ULID
	WorldID     ULID
	Name        string
	Description string
	AuditFields AuditFields
}
