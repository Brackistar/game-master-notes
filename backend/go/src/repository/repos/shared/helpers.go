package shared

import (
	"time"
)

var NowFn = func() time.Time { return time.Now().UTC() }
