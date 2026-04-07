package model

type Player struct {
	ID          ULID
	Name        string
	AuditFields AuditFields
}
