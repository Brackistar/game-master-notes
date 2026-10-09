package model

import "time"

// Links a player to a campaign.
type CampaignPlayer struct {
	CampaignID ULID
	PlayerID   ULID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
