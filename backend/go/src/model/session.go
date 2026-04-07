package model

import "time"

type Session struct {
	ID          ULID
	CampaignID  ULID
	PlayedOn    *time.Time
	SummaryMD   string
	AuditFields AuditFields
}
