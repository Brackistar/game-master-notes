package model

// Labels notes for organization and filtering.
type Tag struct {
	ID          ULID
	Name        string
	CampaignID  *ULID
	AuditFields AuditFields
}
