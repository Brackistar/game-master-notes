package constants

type AssetType int8

const (
	Image AssetType = iota
)

func (a AssetType) String() string {
	return [...]string{"image"}[a]
}

func (a AssetType) EnumIndex() int {
	return int(a)
}
