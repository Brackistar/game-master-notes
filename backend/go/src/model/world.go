package model

import "github.com/Brackistar/game-master-notes/backend/go/src/model/constants"

// Contains the top-level setting for campaigns.
type World struct {
	ID          ULID
	PlaneID     ULID
	Name        string
	Description string
	Status      constants.WorldStatus
	AuditFields AuditFields
}
