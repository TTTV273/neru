//go:build darwin

package darwin

/*
#include <stdlib.h>
#include "inputsource.h"
*/
import "C"

import "unsafe"

// StartInputSourceObserver primes the input source cache and keeps it current.
func StartInputSourceObserver() {
	C.NeruInputSourceStartObserving()
}

// CurrentInputSourceID returns the cached identifier of the selected input
// source, or "" before the cache is primed. It never blocks.
func CurrentInputSourceID() string {
	cID := C.NeruCurrentInputSourceID()
	if cID == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cID))

	return C.GoString(cID)
}

// SelectInputSource queues a switch to the input source with the given
// identifier on the main queue.
func SelectInputSource(id string) {
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))

	C.NeruSelectInputSource(cID)
}
