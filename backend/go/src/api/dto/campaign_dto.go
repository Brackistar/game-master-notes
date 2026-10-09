package dto

import "time"

type CreateCampaignRequest struct {
	WorldID   string  `json:"world_id"`
	Name      string  `json:"name"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

type UpdateCampaignRequest struct {
	Name      string  `json:"name"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

type CampaignResponse struct {
	ID        string     `json:"id"`
	WorldID   string     `json:"world_id"`
	Name      string     `json:"name"`
	StartDate *string    `json:"start_date,omitempty"`
	EndDate   *string    `json:"end_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Version   int32      `json:"version"`
}

type ListCampaignsResponse struct {
	Items          []CampaignResponse `json:"items"`
	Offset         int32              `json:"offset"`
	Limit          int32              `json:"limit"`
	IncludeDeleted bool               `json:"include_deleted"`
}

type CampaignPlayerResponse struct {
	CampaignID string     `json:"campaign_id"`
	PlayerID   string     `json:"player_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type ListCampaignPlayersResponse struct {
	Items          []CampaignPlayerResponse `json:"items"`
	Offset         int32                    `json:"offset"`
	Limit          int32                    `json:"limit"`
	IncludeDeleted bool                     `json:"include_deleted"`
}
