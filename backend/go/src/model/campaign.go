package model

import "time"

// Represents a story arc in a world.
type Campaign struct {
	ID          ULID
	WorldID     ULID
	Name        string
	StartDate   *time.Time
	EndDate     *time.Time
	AuditFields AuditFields
}
