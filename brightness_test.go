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
