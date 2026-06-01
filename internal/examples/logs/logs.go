// Package logs is the pure (HTTP-free) domain for the Live Log Feed example:
// a deterministic synthetic log-line generator. It holds no state and is fully
// unit-testable; the SSE handler supplies wall-clock time per tick.
package logs

import "time"

// Level is a log severity.
type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
)

// String returns the lowercase level name (also the CSS-class suffix).
func (l Level) String() string { return string(l) }

// Severity is the ordering used by the min-level filter (higher = more severe).
func (l Level) Severity() int {
	switch l {
	case Error:
		return 3
	case Warn:
		return 2
	case Info:
		return 1
	default: // Debug
		return 0
	}
}

// LogLine is one rendered feed row. Time is supplied by the caller.
type LogLine struct {
	Time    time.Time
	Level   Level
	Source  string
	Message string
}

// pool is the fixed, deterministic sample of log lines (Time is filled per tick).
var pool = []LogLine{
	{Level: Info, Source: "http", Message: "GET /examples/logs 200 in 4ms"},
	{Level: Debug, Source: "cache", Message: "hit key=session:1f2c ttl=300s"},
	{Level: Info, Source: "auth", Message: "user 4821 signed in"},
	{Level: Warn, Source: "ratelimit", Message: "client 10.0.0.5 nearing quota (92%)"},
	{Level: Info, Source: "worker", Message: "job email.send completed in 120ms"},
	{Level: Error, Source: "db", Message: "connection reset; retrying (attempt 2/3)"},
	{Level: Debug, Source: "router", Message: "matched route /api/examples/logs/stream"},
	{Level: Info, Source: "http", Message: "POST /api/examples/todo/add 200 in 7ms"},
	{Level: Warn, Source: "tls", Message: "certificate for cdn expires in 9 days"},
	{Level: Error, Source: "payments", Message: "charge declined: insufficient_funds"},
	{Level: Info, Source: "worker", Message: "scheduled cleanup removed 12 rows"},
	{Level: Debug, Source: "cache", Message: "evicted 3 cold entries"},
}

// PoolSize is the number of distinct sample lines.
func PoolSize() int { return len(pool) }

// Line returns the i-th synthetic line, cycling the pool. Safe for any int.
func Line(i int) LogLine {
	n := len(pool)
	idx := ((i % n) + n) % n // floored modulo, handles negatives
	return pool[idx]
}
