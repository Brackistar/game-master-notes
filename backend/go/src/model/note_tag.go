package model

import "time"

type NoteTag struct {
	NoteID    ULID
	TagID     ULID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
