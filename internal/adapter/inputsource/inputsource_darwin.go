//go:build darwin

package inputsource

import (
	"github.com/y3owk1n/neru/internal/adapter/platform/darwin"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/ports"
)

// Adapter implements ports.InputSourcePort with macOS TIS input sources.
type Adapter struct{}

// NewAdapter creates the adapter and starts keeping its cache current.
func NewAdapter() *Adapter {
	darwin.StartInputSourceObserver()

	return &Adapter{}
}

// CurrentInputSourceID implements ports.InputSourcePort.
func (a *Adapter) CurrentInputSourceID() string {
	return darwin.CurrentInputSourceID()
}

// SelectInputSource implements ports.InputSourcePort.
func (a *Adapter) SelectInputSource(id string) error {
	if id == "" {
		return derrors.New(derrors.CodeInvalidInput, "input source identifier is empty")
	}

	darwin.SelectInputSource(id)

	return nil
}

var _ ports.InputSourcePort = (*Adapter)(nil)
