package repos

import (
	"fmt"

	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/postgres/generated"
)

func toDBOwnerType(ownerType constants.OwnerType) (generated.OwnerType, error) {
	switch ownerType {
	case constants.World:
		return generated.OwnerTypeWorld, nil
	case constants.Plane:
		return generated.OwnerTypePlane, nil
	case constants.Campaign:
		return generated.OwnerTypeCampaign, nil
	case constants.Session:
		return generated.OwnerTypeSession, nil
	case constants.Player:
		return generated.OwnerTypePlayer, nil
	default:
		return "", fmt.Errorf("invalid owner type enum value: %d", ownerType)
	}
}

func fromDBOwnerType(ownerType generated.OwnerType) (constants.OwnerType, error) {
	switch ownerType {
	case generated.OwnerTypeWorld:
		return constants.World, nil
	case generated.OwnerTypePlane:
		return constants.Plane, nil
	case generated.OwnerTypeCampaign:
		return constants.Campaign, nil
	case generated.OwnerTypeSession:
		return constants.Session, nil
	case generated.OwnerTypePlayer:
		return constants.Player, nil
	default:
		return constants.World, fmt.Errorf("invalid owner type db value: %q", ownerType)
	}
}

func toDBAssetType(assetType constants.AssetType) (generated.AssetType, error) {
	switch assetType {
	case constants.Image:
		return generated.AssetTypeImage, nil
	default:
		return "", fmt.Errorf("invalid asset type enum value: %d", assetType)
	}
}

func fromDBAssetType(assetType generated.AssetType) (constants.AssetType, error) {
	switch assetType {
	case generated.AssetTypeImage:
		return constants.Image, nil
	default:
		return constants.Image, fmt.Errorf("invalid asset type db value: %q", assetType)
	}
}

func toDBNoteLinkType(linkType constants.NoteLinkType) (generated.NoteLinkType, error) {
	switch linkType {
	case constants.Related:
		return generated.NoteLinkTypeRelated, nil
	case constants.Contains:
		return generated.NoteLinkTypeContains, nil
	case constants.Mentions:
		return generated.NoteLinkTypeMentions, nil
	case constants.DependsOn:
		return generated.NoteLinkTypeDependsOn, nil
	case constants.LocatedIn:
		return generated.NoteLinkTypeLocatedIn, nil
	default:
		return "", fmt.Errorf("invalid note link type enum value: %d", linkType)
	}
}

func fromDBNoteLinkType(linkType generated.NoteLinkType) (constants.NoteLinkType, error) {
	switch linkType {
	case generated.NoteLinkTypeRelated:
		return constants.Related, nil
	case generated.NoteLinkTypeContains:
		return constants.Contains, nil
	case generated.NoteLinkTypeMentions:
		return constants.Mentions, nil
	case generated.NoteLinkTypeDependsOn:
		return constants.DependsOn, nil
	case generated.NoteLinkTypeLocatedIn:
		return constants.LocatedIn, nil
	default:
		return constants.Related, fmt.Errorf("invalid note link type db value: %q", linkType)
	}
}
