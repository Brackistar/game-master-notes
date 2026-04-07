package model

import "github.com/Brackistar/game-master-notes/backend/go/src/model/constants"

type World struct {
	ID          ULID
	Name        string
	Description string
	Status      constants.WorldStatus
	AuditFields AuditFields
}
