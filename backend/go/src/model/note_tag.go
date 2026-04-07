package model

import "time"

// Links a note with a tag.
type NoteTag struct {
	NoteID    ULID
	TagID     ULID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
