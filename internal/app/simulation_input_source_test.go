package app_test

import (
	"slices"
	"testing"

	"github.com/y3owk1n/neru/internal/domain"
)

const (
	// simUserInputSource is the input source the fixture desktop starts on: a
	// Vietnamese Telex input method, which composes mode keys into other
	// characters unless a mode switches it out.
	simUserInputSource = "com.apple.inputmethod.VietnameseIM.VietnameseSimpleTelex"
	// simNavigationInputSource is the Latin layout a mode types in.
	simNavigationInputSource = "com.apple.keylayout.ABC"
)

// TestSimulation_ModesTypeInABCAndGiveTheUserInputSourceBack is the journey of
// a user who types Vietnamese with Telex. Opening a mode has to switch to ABC,
// or every hint label and grid key is composed by the input method before the
// mode sees it. Moving straight into another mode must keep ABC up, and only
// the return to idle may hand Telex back.
func TestSimulation_ModesTypeInABCAndGiveTheUserInputSourceBack(t *testing.T) {
	sim := newSimHarness(t, simConfig(), threeButtons(t))

	sim.pressHotkey(hintsHotkey)
	sim.waitMode(domain.ModeHints)

	if got := sim.inputSource.CurrentInputSourceID(); got != simNavigationInputSource {
		t.Fatalf("hints opened on input source %q, want %q", got, simNavigationInputSource)
	}

	sim.pressHotkey(gridHotkey)
	sim.waitMode(domain.ModeGrid)

	if got := sim.inputSource.Selected(); !slices.Equal(got, []string{simNavigationInputSource}) {
		t.Fatalf("switching hints to grid selected %q, want only the one switch to ABC", got)
	}

	sim.press("Escape")
	sim.waitMode(domain.ModeIdle)

	if got := sim.inputSource.CurrentInputSourceID(); got != simUserInputSource {
		t.Fatalf("idle is on input source %q, want the user's own %q back", got, simUserInputSource)
	}

	want := []string{simNavigationInputSource, simUserInputSource}
	if got := sim.inputSource.Selected(); !slices.Equal(got, want) {
		t.Fatalf("input source switches = %q, want %q", got, want)
	}
}
