package model

import "time"

// File created to define in a single place commonly used alias and types

// type alias from string to represent ULID ids on other entities
type ULID string

type Version int32

type AuditFields struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Version   Version
}
