// Package chat holds the pure, HTTP-free domain for the /examples/chat app:
// a deterministic ELIZA bot, the identity cookie, and a RAM-only broadcast hub.
package chat

import (
	"regexp"
	"strings"
)

// rule is one ELIZA pattern: a compiled matcher and a reply template. If the
// template contains "%s", the first capture group is substituted in.
type rule struct {
	re    *regexp.Regexp
	reply string
}

// rules are evaluated in order; the first match wins. Deterministic by design —
// no randomness, so the same input always yields the same reply (E2E relies on it).
var rules = []rule{
	{regexp.MustCompile(`(?i)\bi feel ([a-z ]+)`), "Why do you feel %s?"},
	{regexp.MustCompile(`(?i)\bi am ([a-z ]+)`), "How long have you been %s?"},
	{regexp.MustCompile(`(?i)\bi (?:need|want) ([a-z ]+)`), "Why do you need %s?"},
	{regexp.MustCompile(`(?i)\b(hello|hi|hey)\b`), "Hello. How are you feeling today?"},
	{regexp.MustCompile(`(?i)\b(bye|goodbye)\b`), "Goodbye. Take care of yourself."},
	{regexp.MustCompile(`(?i)\b(because)\b`), "Is that the real reason?"},
	{regexp.MustCompile(`(?i)\b(sorry)\b`), "Please don't apologise."},
	{regexp.MustCompile(`(?i)\byes\b`), "You seem quite sure."},
	{regexp.MustCompile(`(?i)\bno\b`), "Why not?"},
	{regexp.MustCompile(`(?i)\?\s*$`), "Why do you ask?"},
}

// Reply returns a deterministic ELIZA-style response for msg. matched is false
// when no rule fires, so the caller can choose to stay silent.
func Reply(msg string) (reply string, matched bool) {
	trimmed := strings.TrimSpace(msg)
	for _, r := range rules {
		m := r.re.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		if strings.Contains(r.reply, "%s") && len(m) > 1 {
			arg := strings.TrimSpace(m[1])
			// Trim a trailing question mark the capture may have swallowed.
			arg = strings.TrimRight(arg, "?. ")
			return strings.Replace(r.reply, "%s", arg, 1), true
		}
		return r.reply, true
	}
	return "", false
}
