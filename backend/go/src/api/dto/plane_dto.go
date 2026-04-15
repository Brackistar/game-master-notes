package dto

import "time"

type CreatePlaneRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePlaneRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PlaneResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Version     int32      `json:"version"`
}

type ListPlanesResponse struct {
	Items          []PlaneResponse `json:"items"`
	Offset         int32           `json:"offset"`
	Limit          int32           `json:"limit"`
	IncludeDeleted bool            `json:"include_deleted"`
}
