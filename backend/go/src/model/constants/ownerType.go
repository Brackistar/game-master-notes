package constants

type OwnerType int8

const (
	World OwnerType = iota
	Plane
	Campaign
	Session
	Player
)

func (o OwnerType) String() string {
	return [...]string{"world", "plane", "campaign", "session", "player"}[o]
}

func (o OwnerType) EnumIndex() int {
	return int(o)
}
