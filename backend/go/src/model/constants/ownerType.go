package constants

// Defines which entity owns a note.
type OwnerType uint8

const (
	// Sets a world as the owner.
	World OwnerType = iota
	// Sets a plane as the owner.
	Plane
	// Sets a campaign as the owner.
	Campaign
	// Sets a session as the owner.
	Session
	// Sets a player as the owner.
	Player
)

func (o OwnerType) String() string {
	return [...]string{"world", "plane", "campaign", "session", "player"}[o]
}

func (o OwnerType) EnumIndex() int {
	return int(o)
}
