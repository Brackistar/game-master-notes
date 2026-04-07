package model

import "github.com/Brackistar/game-master-notes/backend/go/src/model/constants"

// Stores markdown content and structured metadata.
type Note struct {
	ID           ULID
	Title        string
	ContentMD    string
	NoteType     constants.NoteType
	MetadataJSON []byte
	AuditFields  AuditFields
}
