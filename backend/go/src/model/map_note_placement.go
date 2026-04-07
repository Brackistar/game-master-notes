package model

// Stores a note marker position on a map note.
type MapNotePlacement struct {
	ID           ULID
	MapNoteID    ULID
	TargetNoteID ULID
	X            uint8
	Y            uint8
	AuditFields  AuditFields
}
