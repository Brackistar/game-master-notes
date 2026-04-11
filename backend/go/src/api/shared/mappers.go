package shared

import (
	"github.com/Brackistar/game-master-notes/backend/go/src/api/dto"
	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

func MapPlayerToDTO(player model.Player) dto.PlayerResponse {
	return dto.PlayerResponse{
		ID:        string(player.ID),
		Name:      player.Name,
		CreatedAt: player.AuditFields.CreatedAt,
		UpdatedAt: player.AuditFields.UpdatedAt,
		DeletedAt: player.AuditFields.DeletedAt,
		Version:   int32(player.AuditFields.Version),
	}
}

func MapCampaignToDTO(campaign model.Campaign) dto.CampaignResponse {
	return dto.CampaignResponse{
		ID:        string(campaign.ID),
		WorldID:   string(campaign.WorldID),
		Name:      campaign.Name,
		StartDate: FormatDatePointer(campaign.StartDate),
		EndDate:   FormatDatePointer(campaign.EndDate),
		CreatedAt: campaign.AuditFields.CreatedAt,
		UpdatedAt: campaign.AuditFields.UpdatedAt,
		DeletedAt: campaign.AuditFields.DeletedAt,
		Version:   int32(campaign.AuditFields.Version),
	}
}

func MapCampaignPlayerToDTO(rel model.CampaignPlayer) dto.CampaignPlayerResponse {
	return dto.CampaignPlayerResponse{
		CampaignID: string(rel.CampaignID),
		PlayerID:   string(rel.PlayerID),
		CreatedAt:  rel.CreatedAt,
		UpdatedAt:  rel.UpdatedAt,
		DeletedAt:  rel.DeletedAt,
	}
}

func MapWorldToDTO(world model.World) dto.WorldResponse {
	return dto.WorldResponse{
		ID:          string(world.ID),
		Name:        world.Name,
		Description: world.Description,
		Status:      world.Status.String(),
		CreatedAt:   world.AuditFields.CreatedAt,
		UpdatedAt:   world.AuditFields.UpdatedAt,
		DeletedAt:   world.AuditFields.DeletedAt,
		Version:     int32(world.AuditFields.Version),
	}
}

func MapPlaneToDTO(plane model.Plane) dto.PlaneResponse {
	return dto.PlaneResponse{
		ID:          string(plane.ID),
		WorldID:     string(plane.WorldID),
		Name:        plane.Name,
		Description: plane.Description,
		CreatedAt:   plane.AuditFields.CreatedAt,
		UpdatedAt:   plane.AuditFields.UpdatedAt,
		DeletedAt:   plane.AuditFields.DeletedAt,
		Version:     int32(plane.AuditFields.Version),
	}
}

func MapSessionToDTO(session model.Session) dto.SessionResponse {
	return dto.SessionResponse{
		ID:         string(session.ID),
		CampaignID: string(session.CampaignID),
		PlayedOn:   FormatDatePointer(session.PlayedOn),
		SummaryMD:  session.SummaryMD,
		CreatedAt:  session.AuditFields.CreatedAt,
		UpdatedAt:  session.AuditFields.UpdatedAt,
		DeletedAt:  session.AuditFields.DeletedAt,
		Version:    int32(session.AuditFields.Version),
	}
}

func MapTagToDTO(tag model.Tag) dto.TagResponse {
	return dto.TagResponse{
		ID:         string(tag.ID),
		Name:       tag.Name,
		CampaignID: ToStringPointer(tag.CampaignID),
		CreatedAt:  tag.AuditFields.CreatedAt,
		UpdatedAt:  tag.AuditFields.UpdatedAt,
		DeletedAt:  tag.AuditFields.DeletedAt,
		Version:    int32(tag.AuditFields.Version),
	}
}

func MapNoteToDTO(note model.Note) dto.NoteResponse {
	return dto.NoteResponse{
		ID:           string(note.ID),
		Title:        note.Title,
		ContentMD:    note.ContentMD,
		NoteType:     note.NoteType.String(),
		MetadataJSON: note.MetadataJSON,
		CreatedAt:    note.AuditFields.CreatedAt,
		UpdatedAt:    note.AuditFields.UpdatedAt,
		DeletedAt:    note.AuditFields.DeletedAt,
		Version:      int32(note.AuditFields.Version),
	}
}

func MapNoteOwnerToDTO(rel model.NoteOwner) dto.NoteOwnerResponse {
	return dto.NoteOwnerResponse{
		NoteID:    string(rel.NoteID),
		OwnerType: rel.OwnerType.String(),
		OwnerID:   string(rel.OwnerID),
		CreatedAt: rel.CreatedAt,
		UpdatedAt: rel.UpdatedAt,
		DeletedAt: rel.DeletedAt,
	}
}

func MapNoteTagToDTO(rel model.NoteTag) dto.NoteTagResponse {
	return dto.NoteTagResponse{
		NoteID:    string(rel.NoteID),
		TagID:     string(rel.TagID),
		CreatedAt: rel.CreatedAt,
		UpdatedAt: rel.UpdatedAt,
		DeletedAt: rel.DeletedAt,
	}
}

func MapNoteLinkToDTO(link model.NoteLink) dto.NoteLinkResponse {
	return dto.NoteLinkResponse{
		ID:           string(link.ID),
		SourceNoteID: string(link.SourceNoteID),
		TargetNoteID: string(link.TargetNoteID),
		LinkType:     link.LinkType.String(),
		CreatedAt:    link.AuditFields.CreatedAt,
		UpdatedAt:    link.AuditFields.UpdatedAt,
		DeletedAt:    link.AuditFields.DeletedAt,
		Version:      int32(link.AuditFields.Version),
	}
}

func MapNoteAssetToDTO(asset model.NoteAsset) dto.NoteAssetResponse {
	return dto.NoteAssetResponse{
		ID:          string(asset.ID),
		NoteID:      string(asset.NoteID),
		AssetType:   asset.AssetType.String(),
		StoragePath: asset.StoragePath,
		MIMEType:    asset.MIMEType,
		CreatedAt:   asset.AuditFields.CreatedAt,
		UpdatedAt:   asset.AuditFields.UpdatedAt,
		DeletedAt:   asset.AuditFields.DeletedAt,
		Version:     int32(asset.AuditFields.Version),
	}
}

func MapMapPlacementToDTO(placement model.MapNotePlacement) dto.MapNotePlacementResponse {
	return dto.MapNotePlacementResponse{
		ID:           string(placement.ID),
		MapNoteID:    string(placement.MapNoteID),
		TargetNoteID: string(placement.TargetNoteID),
		X:            placement.X,
		Y:            placement.Y,
		CreatedAt:    placement.AuditFields.CreatedAt,
		UpdatedAt:    placement.AuditFields.UpdatedAt,
		DeletedAt:    placement.AuditFields.DeletedAt,
		Version:      int32(placement.AuditFields.Version),
	}
}
