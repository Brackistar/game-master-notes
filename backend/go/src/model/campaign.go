package model

import "time"

type Campaign struct {
	ID          ULID
	WorldID     ULID
	Name        string
	StartDate   *time.Time
	EndDate     *time.Time
	AuditFields AuditFields
}
