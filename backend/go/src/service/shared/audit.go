package shared

import (
	"log/slog"
	"runtime"
	"strings"
	"time"
)

func LogServiceCall() func() {
	name := callerName(2)
	start := time.Now().UTC()
	slog.Info("service_call_start", "call", name, "at", start.Format(time.RFC3339Nano))
	return func() {
		slog.Info("service_call_end", "call", name, "duration_ms", time.Since(start).Milliseconds())
	}
}

func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	full := fn.Name()
	if idx := strings.LastIndex(full, "/"); idx >= 0 {
		full = full[idx+1:]
	}
	return full
}
