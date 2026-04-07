package model

import "time"

type CampaignPlayer struct {
	CampaignID ULID
	PlayerID   ULID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
