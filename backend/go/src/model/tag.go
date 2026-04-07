package model

type Tag struct {
	ID          ULID
	Name        string
	CampaignID  *ULID
	AuditFields AuditFields
}
