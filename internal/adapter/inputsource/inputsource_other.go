//go:build !darwin

package inputsource

import (
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/ports"
)

// Adapter is the slot for platforms where Neru does not switch input sources.
type Adapter struct{}

// NewAdapter creates the adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// CurrentInputSourceID reports that the input source is not known, so the
// mode handler leaves it alone.
func (a *Adapter) CurrentInputSourceID() string {
	return ""
}

// SelectInputSource reports that this platform has no input source switching.
func (a *Adapter) SelectInputSource(_ string) error {
	return derrors.New(
		derrors.CodeNotSupported,
		"input source switching is only supported on macOS",
	)
}

var _ ports.InputSourcePort = (*Adapter)(nil)
