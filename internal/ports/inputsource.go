package ports

// InputSourcePort reads and selects the system keyboard input source (on macOS
// a TIS input source such as "com.apple.keylayout.ABC" or a Vietnamese Telex
// input method).
//
// The mode handler calls both methods under h.mu, so neither may block:
// CurrentInputSourceID answers from a cache the platform keeps current, and
// SelectInputSource only queues the switch. A platform without input sources
// answers "" and CodeNotSupported, and the handler then leaves the input
// source alone.
type InputSourcePort interface {
	// CurrentInputSourceID returns the identifier of the selected input
	// source, or "" when it is not known.
	CurrentInputSourceID() string

	// SelectInputSource queues a switch to the input source with the given
	// identifier. A source that turns out not to be installed is reported by
	// the platform's own log, because the switch completes after this returns.
	SelectInputSource(id string) error
}
