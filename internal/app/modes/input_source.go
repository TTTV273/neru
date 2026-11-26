package modes

import (
	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
)

// navigationInputSourceID is the input source a mode types in. Mode keys are
// Latin letters, and an input method such as Vietnamese Telex composes them
// into other characters before the mode ever sees them.
const navigationInputSourceID = "com.apple.keylayout.ABC"

// syncInputSource puts the plain Latin input source up while a mode is open
// and gives the user's own back when Neru returns to idle. Caller must hold
// h.mu.
//
// It is driven from setAppMode, which every mode change passes through, and it
// keys off idle versus not-idle rather than off a mode, so a mode-to-mode
// transition (which passes through idle with switchPending set) neither
// restores nor re-saves. The source saved is the one in use when Neru left
// idle, and only a return to idle hands it back.
//
// Neither port call blocks (ports.InputSourcePort), which is what makes it
// safe under h.mu.
func (h *handlerState) syncInputSource(mode domain.Mode) {
	if h.inputSource == nil {
		return
	}

	if mode != domain.ModeIdle {
		h.selectNavigationInputSource()

		return
	}

	h.restoreInputSource()
}

// selectNavigationInputSource saves the input source in use and switches to
// the navigation one, once per stay outside idle.
func (h *handlerState) selectNavigationInputSource() {
	if h.savedInputSource != "" {
		return
	}

	current := h.inputSource.CurrentInputSourceID()
	if current == "" {
		// Not known, or a platform without input sources. Switching without
		// knowing what to give back would strand the user in ABC.
		return
	}

	h.savedInputSource = current

	if current == navigationInputSourceID {
		return
	}

	err := h.inputSource.SelectInputSource(navigationInputSourceID)
	if err != nil {
		h.savedInputSource = ""
		h.logInputSourceError("Failed to select the navigation input source", err)
	}
}

// restoreInputSource gives back the input source saved on the way out of idle.
func (h *handlerState) restoreInputSource() {
	saved := h.savedInputSource
	if saved == "" {
		return
	}

	h.savedInputSource = ""

	if saved == navigationInputSourceID {
		return
	}

	err := h.inputSource.SelectInputSource(saved)
	if err != nil {
		h.logInputSourceError("Failed to restore the input source", err)
	}
}

func (h *handlerState) logInputSourceError(message string, err error) {
	if derrors.IsNotSupported(err) {
		h.logger.Debug(message, zap.Error(err))

		return
	}

	h.logger.Warn(message, zap.Error(err))
}
