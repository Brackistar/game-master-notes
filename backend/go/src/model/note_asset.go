package model

import "github.com/Brackistar/game-master-notes/backend/go/src/model/constants"

// References a file attached to a note.
type NoteAsset struct {
	ID          ULID
	NoteID      ULID
	AssetType   constants.AssetType
	StoragePath string
	MIMEType    string
	AuditFields AuditFields
}
