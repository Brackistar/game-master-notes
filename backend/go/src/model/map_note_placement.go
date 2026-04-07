package model

// Stores a note marker position on a map note.
type MapNotePlacement struct {
	ID           ULID
	MapNoteID    ULID
	TargetNoteID ULID
	X            float64
	Y            float64
	AuditFields  AuditFields
}
