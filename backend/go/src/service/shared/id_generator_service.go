package shared

import (
	"errors"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/oklog/ulid/v2"
)

type IDGenerator interface {
	NewULID() (model.ULID, error)
}

type OklogULIDGenerator struct{}

func NewOklogULIDGenerator() *OklogULIDGenerator {
	return &OklogULIDGenerator{}
}

func (g *OklogULIDGenerator) NewULID() (model.ULID, error) {
	id := ulid.Make()
	if id.Compare(ulid.ULID{}) == 0 {
		return "", errors.New("failed to generate ULID")
	}
	return model.ULID(id.String()), nil
}
