package model

import "github.com/Brackistar/game-master-notes/backend/go/src/model/constants"

type NoteAsset struct {
	ID          ULID
	NoteID      ULID
	AssetType   constants.AssetType
	StoragePath string
	MIMEType    string
	AuditFields AuditFields
}
