package strategy

// CheckoutStrategy enumerates the strategies for checking out files from the cache
type CheckoutStrategy int

const (
	// LinkStrategy creates read-only symbolic links to files in the cache
	LinkStrategy CheckoutStrategy = iota
	// CopyStrategy creates copies of files in the cache
	CopyStrategy
)

func (strat CheckoutStrategy) String() string {
	return [...]string{"LinkStrategy", "CopyStrategy"}[strat]
}
