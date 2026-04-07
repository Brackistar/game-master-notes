package model

type MapNotePlacement struct {
	ID           ULID
	MapNoteID    ULID
	TargetNoteID ULID
	X            float64
	Y            float64
	AuditFields  AuditFields
}
