package error

// GMN_* tokens are a stable contract emitted by PostgreSQL fn_* functions
// (migration 002_service_functions) and consumed by repository/service error mapping.
const (
	GMNCampaignNotFound          = "GMN_CAMPAIGN_NOT_FOUND"
	GMNCampaignDeleted           = "GMN_CAMPAIGN_DELETED"
	GMNPlayerNotFound            = "GMN_PLAYER_NOT_FOUND"
	GMNPlayerDeleted             = "GMN_PLAYER_DELETED"
	GMNCampaignPlayerAlreadyOpen = "GMN_CAMPAIGN_PLAYER_ALREADY_ACTIVE"
	GMNCampaignPlayerNotActive   = "GMN_CAMPAIGN_PLAYER_NOT_ACTIVE"

	GMNNoteNotFound  = "GMN_NOTE_NOT_FOUND"
	GMNNoteDeleted   = "GMN_NOTE_DELETED"
	GMNTagNotFound   = "GMN_TAG_NOT_FOUND"
	GMNTagDeleted    = "GMN_TAG_DELETED"
	GMNMapNoteNotFound = "GMN_MAP_NOTE_NOT_FOUND"
	GMNMapNoteDeleted  = "GMN_MAP_NOTE_DELETED"

	GMNOwnerWorldNotFound    = "GMN_OWNER_NOT_FOUND_WORLD"
	GMNOwnerWorldDeleted     = "GMN_OWNER_DELETED_WORLD"
	GMNOwnerPlaneNotFound    = "GMN_OWNER_NOT_FOUND_PLANE"
	GMNOwnerPlaneDeleted     = "GMN_OWNER_DELETED_PLANE"
	GMNOwnerCampaignNotFound = "GMN_OWNER_NOT_FOUND_CAMPAIGN"
	GMNOwnerCampaignDeleted  = "GMN_OWNER_DELETED_CAMPAIGN"
	GMNOwnerSessionNotFound  = "GMN_OWNER_NOT_FOUND_SESSION"
	GMNOwnerSessionDeleted   = "GMN_OWNER_DELETED_SESSION"
	GMNOwnerPlayerNotFound   = "GMN_OWNER_NOT_FOUND_PLAYER"
	GMNOwnerPlayerDeleted    = "GMN_OWNER_DELETED_PLAYER"

	GMNNoteOwnerAlreadyOpen = "GMN_NOTE_OWNER_ALREADY_ACTIVE"
	GMNNoteOwnerNotActive   = "GMN_NOTE_OWNER_NOT_ACTIVE"
	GMNNoteTagAlreadyOpen   = "GMN_NOTE_TAG_ALREADY_ACTIVE"
	GMNNoteTagNotActive     = "GMN_NOTE_TAG_NOT_ACTIVE"

	GMNNoteLinkSelfReference = "GMN_NOTE_LINK_SELF_REFERENCE"
	GMNSourceNoteNotFound    = "GMN_SOURCE_NOTE_NOT_FOUND"
	GMNSourceNoteDeleted     = "GMN_SOURCE_NOTE_DELETED"
	GMNTargetNoteNotFound    = "GMN_TARGET_NOTE_NOT_FOUND"
	GMNTargetNoteDeleted     = "GMN_TARGET_NOTE_DELETED"
	GMNNoteLinkAlreadyOpen   = "GMN_NOTE_LINK_ALREADY_ACTIVE"
	GMNNoteLinkNotActive     = "GMN_NOTE_LINK_NOT_ACTIVE"

	GMNMapPlacementXOutOfRange = "GMN_MAP_NOTE_PLACEMENT_X_OUT_OF_RANGE"
	GMNMapPlacementYOutOfRange = "GMN_MAP_NOTE_PLACEMENT_Y_OUT_OF_RANGE"
	GMNMapPlacementNotActive   = "GMN_MAP_NOTE_PLACEMENT_NOT_ACTIVE"
)

var allGMNTokens = []string{
	GMNCampaignNotFound,
	GMNCampaignDeleted,
	GMNPlayerNotFound,
	GMNPlayerDeleted,
	GMNCampaignPlayerAlreadyOpen,
	GMNCampaignPlayerNotActive,
	GMNNoteNotFound,
	GMNNoteDeleted,
	GMNTagNotFound,
	GMNTagDeleted,
	GMNMapNoteNotFound,
	GMNMapNoteDeleted,
	GMNOwnerWorldNotFound,
	GMNOwnerWorldDeleted,
	GMNOwnerPlaneNotFound,
	GMNOwnerPlaneDeleted,
	GMNOwnerCampaignNotFound,
	GMNOwnerCampaignDeleted,
	GMNOwnerSessionNotFound,
	GMNOwnerSessionDeleted,
	GMNOwnerPlayerNotFound,
	GMNOwnerPlayerDeleted,
	GMNNoteOwnerAlreadyOpen,
	GMNNoteOwnerNotActive,
	GMNNoteTagAlreadyOpen,
	GMNNoteTagNotActive,
	GMNNoteLinkSelfReference,
	GMNSourceNoteNotFound,
	GMNSourceNoteDeleted,
	GMNTargetNoteNotFound,
	GMNTargetNoteDeleted,
	GMNNoteLinkAlreadyOpen,
	GMNNoteLinkNotActive,
	GMNMapPlacementXOutOfRange,
	GMNMapPlacementYOutOfRange,
	GMNMapPlacementNotActive,
}

func GMNTokens() []string {
	out := make([]string, len(allGMNTokens))
	copy(out, allGMNTokens)
	return out
}
