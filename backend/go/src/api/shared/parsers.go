package shared

import (
	"fmt"

	apierrors "github.com/Brackistar/game-master-notes/backend/go/src/api/error"
	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

func ParseWorldStatus(value string) (constants.WorldStatus, error) {
	switch value {
	case constants.Draft.String():
		return constants.Draft, nil
	case constants.Active.String():
		return constants.Active, nil
	case constants.Archived.String():
		return constants.Archived, nil
	default:
		return constants.Draft, fmt.Errorf(apierrors.APIINVALIDREQUESTMESSAGE)
	}
}

func ParseNoteType(value string) (constants.NoteType, error) {
	switch value {
	case constants.General.String():
		return constants.General, nil
	case constants.SummaryNote.String():
		return constants.SummaryNote, nil
	case constants.Map.String():
		return constants.Map, nil
	case constants.Character.String():
		return constants.Character, nil
	case constants.Location.String():
		return constants.Location, nil
	default:
		return constants.General, fmt.Errorf(apierrors.APIINVALIDREQUESTMESSAGE)
	}
}

func ParseOwnerType(value string) (constants.OwnerType, error) {
	switch value {
	case constants.World.String():
		return constants.World, nil
	case constants.Plane.String():
		return constants.Plane, nil
	case constants.Campaign.String():
		return constants.Campaign, nil
	case constants.Session.String():
		return constants.Session, nil
	case constants.Player.String():
		return constants.Player, nil
	default:
		return constants.World, fmt.Errorf(apierrors.APIINVALIDREQUESTMESSAGE)
	}
}

func ParseNoteLinkType(value string) (constants.NoteLinkType, error) {
	switch value {
	case constants.Related.String():
		return constants.Related, nil
	case constants.Contains.String():
		return constants.Contains, nil
	case constants.Mentions.String():
		return constants.Mentions, nil
	case constants.DependsOn.String():
		return constants.DependsOn, nil
	case constants.LocatedIn.String():
		return constants.LocatedIn, nil
	default:
		return constants.Related, fmt.Errorf(apierrors.APIINVALIDREQUESTMESSAGE)
	}
}

func ParseAssetType(value string) (constants.AssetType, error) {
	switch value {
	case constants.Image.String():
		return constants.Image, nil
	default:
		return constants.Image, fmt.Errorf(apierrors.APIINVALIDREQUESTMESSAGE)
	}
}

func ToStringPointer(value *model.ULID) *string {
	if value == nil {
		return nil
	}
	s := string(*value)
	return &s
}

func FromStringPointer(value *string) *model.ULID {
	if value == nil {
		return nil
	}
	id := model.ULID(*value)
	return &id
}
