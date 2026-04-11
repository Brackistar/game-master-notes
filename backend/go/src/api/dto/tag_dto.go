package dto

import "time"

type CreateTagRequest struct {
	Name       string  `json:"name"`
	CampaignID *string `json:"campaign_id"`
}

type UpdateTagRequest struct {
	Name       string  `json:"name"`
	CampaignID *string `json:"campaign_id"`
}

type TagResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CampaignID *string    `json:"campaign_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	Version    int32      `json:"version"`
}

type ListTagsResponse struct {
	Items          []TagResponse `json:"items"`
	Offset         int32         `json:"offset"`
	Limit          int32         `json:"limit"`
	IncludeDeleted bool          `json:"include_deleted"`
}
