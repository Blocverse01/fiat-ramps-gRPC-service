package ycstatus

import (
	"fmt"
	"strings"
)

const (
	Recorded   = "recorded"
	Sent       = "sent"
	Pending    = "pending"
	Processing = "processing"
	Complete   = "complete"
	Failed     = "failed"
	Expired    = "expired"
)

var allowedStatuses = map[string]struct{}{
	Recorded:   {},
	Sent:       {},
	Pending:    {},
	Processing: {},
	Complete:   {},
	Failed:     {},
	Expired:    {},
}

var terminalStatuses = map[string]struct{}{
	Complete: {},
	Failed:   {},
	Expired:  {},
}

var allowedTransitions = map[string]map[string]struct{}{
	Recorded:   {Sent: {}, Pending: {}, Processing: {}, Complete: {}, Failed: {}, Expired: {}},
	Sent:       {Pending: {}, Processing: {}, Complete: {}, Failed: {}, Expired: {}},
	Pending:    {Processing: {}, Complete: {}, Failed: {}, Expired: {}},
	Processing: {Complete: {}, Failed: {}, Expired: {}},
	Complete:   {},
	Failed:     {},
	Expired:    {},
}

func Normalize(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func IsAllowed(status string) bool {
	_, ok := allowedStatuses[Normalize(status)]
	return ok
}

func IsTerminal(status string) bool {
	_, ok := terminalStatuses[Normalize(status)]
	return ok
}

func IsSentOrComplete(status string) bool {
	s := Normalize(status)
	return s == Sent || s == Complete
}

func CanTransition(from, to string) bool {
	from, to = Normalize(from), Normalize(to)
	if from == to {
		return true
	}
	next, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	_, ok = next[to]
	return ok
}

// ValidateStatusWrite enforces the allowlist, terminal immutability, and that
// the caller status matches the re-fetched provider status.
func ValidateStatusWrite(current, requested, provider string) error {
	requested = Normalize(requested)
	provider = Normalize(provider)
	current = Normalize(current)

	if !IsAllowed(requested) {
		return fmt.Errorf("unknown status %q", requested)
	}
	if provider == "" {
		return fmt.Errorf("provider status is required")
	}
	if !IsAllowed(provider) {
		return fmt.Errorf("unknown provider status %q", provider)
	}
	if requested != provider {
		return fmt.Errorf("requested status %q does not match provider status %q", requested, provider)
	}
	if current != "" && IsTerminal(current) && requested != current {
		return fmt.Errorf("terminal status %q is immutable", current)
	}
	if current != "" && !CanTransition(current, requested) {
		return fmt.Errorf("invalid status transition from %q to %q", current, requested)
	}
	return nil
}
