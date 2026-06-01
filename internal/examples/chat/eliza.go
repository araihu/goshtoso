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

// fallbacks are generic deflections used when no rule matches, so ELIZA always
// says something (like the original). Picked deterministically by message length.
var fallbacks = []string{
	"Tell me more.",
	"How does that make you feel?",
	"Why do you say that?",
	"I see. Please go on.",
	"Can you elaborate on that?",
	"What does that suggest to you?",
}

// Reply returns a deterministic ELIZA-style response for msg. matched is false
// only for empty input; any non-empty message gets either a rule reply or a
// generic fallback, so the bot is never silent when enabled.
func Reply(msg string) (reply string, matched bool) {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return "", false
	}
	for _, r := range rules {
		m := r.re.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		if strings.Contains(r.reply, "%s") && len(m) > 1 {
			arg := strings.TrimSpace(m[1])
			// Trim a trailing question mark the capture may have swallowed.
			arg = strings.TrimRight(arg, "?. ")
			// Skip this rule if the capture is empty — avoids bare-placeholder replies.
			if arg == "" {
				continue
			}
			return strings.Replace(r.reply, "%s", arg, 1), true
		}
		return r.reply, true
	}
	// No rule matched: deflect deterministically rather than stay silent.
	return fallbacks[len(trimmed)%len(fallbacks)], true
}
