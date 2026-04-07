package model

import "github.com/Brackistar/game-master-notes/backend/go/src/model/constants"

type Note struct {
	ID           ULID
	Title        string
	ContentMD    string
	NoteType     constants.NoteType
	MetadataJSON []byte
	AuditFields  AuditFields
}
