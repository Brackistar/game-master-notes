package error

const (
	SERVDEPNILMESSAGE                = "%s service dependency is nil: %s"
	SERVFIELDREQUIREDMESSAGE         = "%s is required"
	SERVFIELDCANNOTBEEMPTYMESSAGE    = "%s cannot be empty"
	SERVOFFSETGTEZEROMESSAGE         = "offset must be >= 0"
	SERVLIMITGTZEROMESSAGE           = "limit must be > 0"
	SERVEXPECTEDVERSIONGTZEROMESSAGE = "expected_version must be > 0"
	SERVINVALIDFIELDMESSAGE          = "invalid %s"
	SERVFIELDMINCHARSMESSAGE         = "%s must be at least %d characters"
	SERVFIELDMAXCHARSMESSAGE         = "%s must be at most %d characters"
	SERVNAMEUNSUPPORTEDCHARSMESSAGE  = "name contains unsupported characters"
	SERVMETADATAJSONVALIDMESSAGE     = "metadata_json must be valid JSON"
	SERVFIELDSMUSTBEDIFFERENTMESSAGE = "%s and %s must be different"
	SERVENDDATEGTESTARTDATEMESSAGE   = "end_date must be greater than or equal to start_date"
	SERVXYMAXRANGEMESSAGE            = "x and y must be <= 100"
	SERVULIDGENFAILEDMESSAGE         = "failed to generate ULID"
)
