package dto

import "time"

type CreateSessionRequest struct {
	CampaignID string  `json:"campaign_id"`
	PlayedOn   *string `json:"played_on"`
	SummaryMD  string  `json:"summary_md"`
}

type UpdateSessionRequest struct {
	PlayedOn  *string `json:"played_on"`
	SummaryMD string  `json:"summary_md"`
}

type SessionResponse struct {
	ID         string     `json:"id"`
	CampaignID string     `json:"campaign_id"`
	PlayedOn   *string    `json:"played_on,omitempty"`
	SummaryMD  string     `json:"summary_md"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	Version    int32      `json:"version"`
}

type ListSessionsResponse struct {
	Items          []SessionResponse `json:"items"`
	Offset         int32             `json:"offset"`
	Limit          int32             `json:"limit"`
	IncludeDeleted bool              `json:"include_deleted"`
}
