package modes

import (
	"slices"
	"testing"

	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/ports/mocks"
)

const testTelexInputSource = "com.apple.inputmethod.VietnameseIM.VietnameseSimpleTelex"

func newInputSourceHandler(port *mocks.MockInputSourcePort) *Handler {
	return newHandlerWithState(handlerState{
		appState:    state.NewAppState(),
		inputSource: port,
	})
}

func TestHandlerState_SyncInputSource(t *testing.T) {
	notSupported := derrors.New(derrors.CodeNotSupported, "no input sources here")

	tests := []struct {
		name      string
		start     string
		selectErr error
		steps     []domain.Mode
		want      []string
		wantFinal string
	}{
		{
			name:      "entering a mode switches to ABC and idle gives the user's back",
			start:     testTelexInputSource,
			steps:     []domain.Mode{domain.ModeHints, domain.ModeIdle},
			want:      []string{navigationInputSourceID, testTelexInputSource},
			wantFinal: testTelexInputSource,
		},
		{
			name:      "a second mode before idle does not re-save ABC as the user's",
			start:     testTelexInputSource,
			steps:     []domain.Mode{domain.ModeHints, domain.ModeGrid, domain.ModeIdle},
			want:      []string{navigationInputSourceID, testTelexInputSource},
			wantFinal: testTelexInputSource,
		},
		{
			name:      "a user already on ABC is never switched",
			start:     navigationInputSourceID,
			steps:     []domain.Mode{domain.ModeHints, domain.ModeIdle},
			want:      nil,
			wantFinal: navigationInputSourceID,
		},
		{
			name:      "an unknown input source is left alone",
			start:     "",
			steps:     []domain.Mode{domain.ModeHints, domain.ModeIdle},
			want:      nil,
			wantFinal: "",
		},
		{
			name:      "idle without a mode before it restores nothing",
			start:     testTelexInputSource,
			steps:     []domain.Mode{domain.ModeIdle},
			want:      nil,
			wantFinal: testTelexInputSource,
		},
		{
			name:      "a failed switch saves nothing, so idle restores nothing",
			start:     testTelexInputSource,
			selectErr: notSupported,
			steps:     []domain.Mode{domain.ModeHints, domain.ModeIdle},
			want:      []string{navigationInputSourceID},
			wantFinal: testTelexInputSource,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			port := mocks.NewMockInputSourcePort(testCase.start)
			if testCase.selectErr != nil {
				port.SelectFunc = func(string) error { return testCase.selectErr }
			}

			handler := newInputSourceHandler(port)

			for _, mode := range testCase.steps {
				handler.syncInputSource(mode)
			}

			if got := port.Selected(); !slices.Equal(got, testCase.want) {
				t.Errorf("selected = %q, want %q", got, testCase.want)
			}

			if got := port.CurrentInputSourceID(); got != testCase.wantFinal {
				t.Errorf("final input source = %q, want %q", got, testCase.wantFinal)
			}

			if handler.savedInputSource != "" {
				t.Errorf("saved input source %q outlived the return to idle", handler.savedInputSource)
			}
		})
	}
}

func TestHandlerState_SyncInputSource_NilPortDoesNothing(t *testing.T) {
	handler := newHandlerWithState(handlerState{appState: state.NewAppState()})

	handler.syncInputSource(domain.ModeHints)
	handler.syncInputSource(domain.ModeIdle)

	if handler.savedInputSource != "" {
		t.Errorf("saved input source = %q with no port, want none", handler.savedInputSource)
	}
}

// An activation that exits one mode for another and then gives up lands in
// idle through releaseKeyboardIfNoModeEntered, never through a setAppMode the
// transition let restore. That path has to hand the user's input source back.
func TestHandlerState_ReleaseKeyboardIfNoModeEntered_RestoresTheInputSource(t *testing.T) {
	port := mocks.NewMockInputSourcePort(testTelexInputSource)
	handler := newInputSourceHandler(port)

	handler.syncInputSource(domain.ModeHints)
	handler.releaseKeyboardIfNoModeEntered()

	if got := port.CurrentInputSourceID(); got != testTelexInputSource {
		t.Errorf("input source after an abandoned transition = %q, want %q", got, testTelexInputSource)
	}
}
