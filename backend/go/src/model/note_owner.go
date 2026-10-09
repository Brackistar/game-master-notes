package model

import (
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

// Assigns a note to an owning entity.
type NoteOwner struct {
	NoteID    ULID
	OwnerType constants.OwnerType
	OwnerID   ULID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
