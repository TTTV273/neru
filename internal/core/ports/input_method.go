package ports

// InputMethodPort defines the interface for reading and switching the system input source.
type InputMethodPort interface {
	// GetCurrentSourceID returns the identifier of the currently active input source
	// (e.g. "com.apple.keylayout.ABC").
	GetCurrentSourceID() string

	// SwitchToSourceID activates the input source with the given identifier.
	// Returns an error if the source cannot be found or selected.
	SwitchToSourceID(id string) error
}
