package model

import "github.com/Brackistar/game-master-notes/backend/go/src/model/constants"

type NoteLink struct {
	ID           ULID
	SourceNoteID ULID
	TargetNoteID ULID
	LinkType     constants.NoteLinkType
	AuditFields  AuditFields
}
