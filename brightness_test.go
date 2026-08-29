// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

package brightness

import (
	"errors"
	"testing"
)

// TestSetRefusesALevelOutsideTheRange before it asks the system anything: a
// panel has no brightness below off or above full, and saying so costs no
// round-trip.
func TestSetRefusesALevelOutsideTheRange(t *testing.T) {
	for _, level := range []float64{-0.01, -1, 1.01, 2} {
		if err := Set(1, level); !errors.Is(err, ErrRange) {
			t.Errorf("Set(_, %v) = %v, want ErrRange", level, err)
		}
	}
}

// TestTheEdgesOfTheRangeAreLevels: off and full are legal and must not be
// refused by an off-by-one.
func TestTheEdgesOfTheRangeAreLevels(t *testing.T) {
	for _, level := range []float64{Off, Full} {
		if err := Set(0, level); errors.Is(err, ErrRange) {
			t.Errorf("Set(_, %v) was refused for its level", level)
		}
	}
}

// TestEveryEntryPointAnswersForADisplayThatIsNotThere.
//
// It uses an id no machine has, on purpose: this test runs on darwin too, where
// Dim on a REAL display would turn somebody's screen off in the middle of a
// test run. What it pins is that every entry point answers rather than
// panicking, and that Dim hands back no restore when there was nothing to
// change.
func TestEveryEntryPointAnswersForADisplayThatIsNotThere(t *testing.T) {
	const noSuchDisplay = ^uint32(0)

	if _, err := Of(noSuchDisplay); err == nil {
		t.Error("a display that does not exist reported a brightness")
	}
	restore, err := Dim(noSuchDisplay)
	if err == nil {
		t.Error("a display that does not exist was dimmed")
	}
	if restore != nil {
		t.Error("Dim handed back a way home for something it never changed")
	}
	// And the error says which of the two things went wrong, whichever
	// platform this is.
	if !errors.Is(err, ErrUnsupported) && !errors.Is(err, ErrNoBrightness) &&
		!errors.Is(err, ErrUnavailable) {
		t.Errorf("Dim = %v, want one of this package's errors", err)
	}
}

// swap installs a fake backlight for the duration of one test and puts the
// real seams back afterwards.
func swap(t *testing.T, g func(uint32) (float64, error), s func(uint32, float64) error) {
	t.Helper()
	oldGet, oldSet := get, set
	get, set = g, s
	t.Cleanup(func() { get, set = oldGet, oldSet })
}

func TestDimGoesBackToExactlyWhatItFound(t *testing.T) {
	const display = 1
	panel := 0.43
	swap(t,
		func(uint32) (float64, error) { return panel, nil },
		func(_ uint32, level float64) error { panel = level; return nil },
	)

	restore, err := Dim(display)
	if err != nil {
		t.Fatalf("Dim: %v", err)
	}
	if panel != Off {
		t.Errorf("after Dim the panel is at %v, want %v", panel, Off)
	}
	if err := restore(); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if panel != 0.43 {
		t.Errorf("restored to %v, want the 0.43 it found", panel)
	}
}

func TestDimReportsAPanelThatWillNotBeTurnedDown(t *testing.T) {
	stuck := errors.New("the panel refused")
	swap(t,
		func(uint32) (float64, error) { return 0.5, nil },
		func(uint32, float64) error { return stuck },
	)

	restore, err := Dim(1)
	if !errors.Is(err, stuck) {
		t.Errorf("Dim = %v, want the refusal from the panel", err)
	}
	if restore != nil {
		// The negative control for the test above: a way home is only
		// handed back when there is something to go home from.
		t.Error("Dim handed back a way home after changing nothing")
	}
}
