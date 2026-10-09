package dto

import "time"

type CreateWorldRequest struct {
	PlaneID     string `json:"plane_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type UpdateWorldRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type WorldResponse struct {
	ID          string     `json:"id"`
	PlaneID     string     `json:"plane_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Version     int32      `json:"version"`
}

type ListWorldsResponse struct {
	Items          []WorldResponse `json:"items"`
	Offset         int32           `json:"offset"`
	Limit          int32           `json:"limit"`
	IncludeDeleted bool            `json:"include_deleted"`
}
