package model

import (
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

type NoteOwner struct {
	NoteID    ULID
	OwnerType constants.OwnerType
	OwnerID   ULID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
