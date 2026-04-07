package constants

// Defines the kind of asset attached to a note.
type AssetType int8

const (
	// Represent an image file asset.
	Image AssetType = iota
)

func (a AssetType) String() string {
	return [...]string{"image"}[a]
}

func (a AssetType) EnumIndex() int {
	return int(a)
}
