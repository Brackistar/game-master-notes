package dto

import "time"

type CreateNoteRequest struct {
	Title        string `json:"title"`
	ContentMD    string `json:"content_md"`
	NoteType     string `json:"note_type"`
	MetadataJSON []byte `json:"metadata_json"`
}

type UpdateNoteRequest struct {
	Title        string `json:"title"`
	ContentMD    string `json:"content_md"`
	NoteType     string `json:"note_type"`
	MetadataJSON []byte `json:"metadata_json"`
}

type NoteResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	ContentMD    string     `json:"content_md"`
	NoteType     string     `json:"note_type"`
	MetadataJSON []byte     `json:"metadata_json"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Version      int32      `json:"version"`
}

type ListNotesResponse struct {
	Items          []NoteResponse `json:"items"`
	Offset         int32          `json:"offset"`
	Limit          int32          `json:"limit"`
	IncludeDeleted bool           `json:"include_deleted"`
}

type AddNoteOwnerRequest struct {
	OwnerType string `json:"owner_type"`
	OwnerID   string `json:"owner_id"`
}

type NoteOwnerResponse struct {
	NoteID    string     `json:"note_id"`
	OwnerType string     `json:"owner_type"`
	OwnerID   string     `json:"owner_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type ListNoteOwnersResponse struct {
	Items          []NoteOwnerResponse `json:"items"`
	Offset         int32               `json:"offset"`
	Limit          int32               `json:"limit"`
	IncludeDeleted bool                `json:"include_deleted"`
}

type NoteTagResponse struct {
	NoteID    string     `json:"note_id"`
	TagID     string     `json:"tag_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type ListNoteTagsResponse struct {
	Items          []NoteTagResponse `json:"items"`
	Offset         int32             `json:"offset"`
	Limit          int32             `json:"limit"`
	IncludeDeleted bool              `json:"include_deleted"`
}

type CreateNoteLinkRequest struct {
	SourceNoteID string `json:"source_note_id"`
	TargetNoteID string `json:"target_note_id"`
	LinkType     string `json:"link_type"`
}

type UpdateNoteLinkRequest struct {
	LinkType string `json:"link_type"`
}

type NoteLinkResponse struct {
	ID           string     `json:"id"`
	SourceNoteID string     `json:"source_note_id"`
	TargetNoteID string     `json:"target_note_id"`
	LinkType     string     `json:"link_type"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Version      int32      `json:"version"`
}

type ListNoteLinksResponse struct {
	Items          []NoteLinkResponse `json:"items"`
	Offset         int32              `json:"offset"`
	Limit          int32              `json:"limit"`
	IncludeDeleted bool               `json:"include_deleted"`
}

type CreateNoteAssetRequest struct {
	AssetType   string `json:"asset_type"`
	StoragePath string `json:"storage_path"`
	MIMEType    string `json:"mime_type"`
}

type UpdateNoteAssetRequest struct {
	AssetType   string `json:"asset_type"`
	StoragePath string `json:"storage_path"`
	MIMEType    string `json:"mime_type"`
}

type NoteAssetResponse struct {
	ID          string     `json:"id"`
	NoteID      string     `json:"note_id"`
	AssetType   string     `json:"asset_type"`
	StoragePath string     `json:"storage_path"`
	MIMEType    string     `json:"mime_type"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Version     int32      `json:"version"`
}

type ListNoteAssetsResponse struct {
	Items          []NoteAssetResponse `json:"items"`
	Offset         int32               `json:"offset"`
	Limit          int32               `json:"limit"`
	IncludeDeleted bool                `json:"include_deleted"`
}

type UpsertMapNotePlacementRequest struct {
	MapNoteID    string `json:"map_note_id"`
	TargetNoteID string `json:"target_note_id"`
	X            uint8  `json:"x"`
	Y            uint8  `json:"y"`
}

type UpdateMapNotePlacementRequest struct {
	X uint8 `json:"x"`
	Y uint8 `json:"y"`
}

type MapNotePlacementResponse struct {
	ID           string     `json:"id"`
	MapNoteID    string     `json:"map_note_id"`
	TargetNoteID string     `json:"target_note_id"`
	X            uint8      `json:"x"`
	Y            uint8      `json:"y"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Version      int32      `json:"version"`
}

type ListMapNotePlacementsResponse struct {
	Items          []MapNotePlacementResponse `json:"items"`
	Offset         int32                      `json:"offset"`
	Limit          int32                      `json:"limit"`
	IncludeDeleted bool                       `json:"include_deleted"`
}
