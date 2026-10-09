package model

import "time"

// Records one played game session in a campaign.
type Session struct {
	ID          ULID
	CampaignID  ULID
	PlayedOn    *time.Time
	SummaryMD   string
	AuditFields AuditFields
}
