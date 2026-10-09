package dto

import "time"

type CreatePlayerRequest struct {
	Name string `json:"name"`
}

type UpdatePlayerRequest struct {
	Name string `json:"name"`
}

type PlayerResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Version   int32      `json:"version"`
}

type ListPlayersResponse struct {
	Items          []PlayerResponse `json:"items"`
	Offset         int32            `json:"offset"`
	Limit          int32            `json:"limit"`
	IncludeDeleted bool             `json:"include_deleted"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
