package model

type Plane struct {
	ID          ULID
	WorldID     ULID
	Name        string
	Description string
	AuditFields AuditFields
}
