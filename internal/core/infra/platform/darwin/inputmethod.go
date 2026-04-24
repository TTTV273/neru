//go:build darwin

package darwin

/*
#include "inputmethod.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"

	derrors "github.com/y3owk1n/neru/internal/core/errors"
	"github.com/y3owk1n/neru/internal/core/ports"
)

// Compile-time check: InputMethodAdapter must implement ports.InputMethodPort.
var _ ports.InputMethodPort = (*InputMethodAdapter)(nil)

// InputMethodAdapter implements ports.InputMethodPort for macOS using Carbon TIS APIs.
type InputMethodAdapter struct{}

// NewInputMethodAdapter creates a new InputMethodAdapter.
func NewInputMethodAdapter() *InputMethodAdapter {
	return &InputMethodAdapter{}
}

// GetCurrentSourceID returns the identifier of the currently active input source.
func (a *InputMethodAdapter) GetCurrentSourceID() string {
	ptr := C.getInputSourceID()
	if ptr == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(ptr))
	return C.GoString(ptr)
}

// SwitchToSourceID activates the input source with the given identifier.
func (a *InputMethodAdapter) SwitchToSourceID(id string) error {
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))
	if C.switchInputSourceByID(cID) != 0 {
		return derrors.New(derrors.CodeNotSupported, "input source not found: "+id)
	}
	return nil
}
