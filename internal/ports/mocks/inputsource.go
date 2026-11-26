package mocks

import (
	"sync"

	"github.com/y3owk1n/neru/internal/ports"
)

// MockInputSourcePort is a mock implementation of ports.InputSourcePort. It
// behaves like a platform with one selected input source: SelectInputSource
// makes the source it was given the current one, unless SelectFunc fails it.
type MockInputSourcePort struct {
	// SelectFunc answers SelectInputSource. Nil accepts every switch.
	SelectFunc func(id string) error

	mu       sync.Mutex
	current  string
	selected []string
}

// NewMockInputSourcePort returns a mock whose current input source is current.
func NewMockInputSourcePort(current string) *MockInputSourcePort {
	return &MockInputSourcePort{current: current}
}

// SetCurrent changes the current input source, as a user switching it by hand.
func (m *MockInputSourcePort) SetCurrent(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.current = id
}

// CurrentInputSourceID implements ports.InputSourcePort.
func (m *MockInputSourcePort) CurrentInputSourceID() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.current
}

// SelectInputSource implements ports.InputSourcePort.
func (m *MockInputSourcePort) SelectInputSource(id string) error {
	m.mu.Lock()
	m.selected = append(m.selected, id)
	m.mu.Unlock()

	if m.SelectFunc != nil {
		err := m.SelectFunc(id)
		if err != nil {
			return err
		}
	}

	m.SetCurrent(id)

	return nil
}

// Selected returns every identifier passed to SelectInputSource, in order.
func (m *MockInputSourcePort) Selected() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]string(nil), m.selected...)
}

// Ensure MockInputSourcePort implements ports.InputSourcePort.
var _ ports.InputSourcePort = (*MockInputSourcePort)(nil)
