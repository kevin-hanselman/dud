package strategy

import (
	"fmt"
	"strings"
)

// CheckoutStrategy enumerates the strategies for checking out files from the cache
type CheckoutStrategy int

const (
	// LinkStrategy creates read-only links to files in the cache (prefers hard links to symbolic)
	LinkStrategy CheckoutStrategy = iota
	// CopyStrategy creates copies of files in the cache
	CopyStrategy
)

func (strat CheckoutStrategy) String() string {
	return [...]string{"LinkStrategy", "CopyStrategy"}[strat]
}

// FromString parses a CheckoutStrategy from a string.
// If parsing fails, FromString returns an error.
// TODO: We might not even need this function. Keeping it for now just because
// it's small.
func FromString(s string) (CheckoutStrategy, error) {
	s = strings.ToLower(s)
	if s == "" || s == "link" {
		return LinkStrategy, nil
	} else if s == "copy" {
		return CopyStrategy, nil
	} else {
		return 0, fmt.Errorf("unable to parse strategy from %v", s)
	}
}
