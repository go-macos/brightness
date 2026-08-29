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
