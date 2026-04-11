package shared

import (
	"fmt"
	"strings"
	"time"

	apierrors "github.com/Brackistar/game-master-notes/backend/go/src/api/error"
)

const campaignDateFormat = "2006-01-02"

func ParseDatePointer(value *string, fieldName string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse(campaignDateFormat, trimmed)
	if err != nil {
		return nil, fmt.Errorf(apierrors.APIFIELDVALIDDATE, fieldName)
	}
	utc := parsed.UTC()
	return &utc, nil
}

func FormatDatePointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(campaignDateFormat)
	return &formatted
}
