// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin

package brightness

import (
	"errors"
	"os"
	"sync"
	"testing"
)

// resetLoad puts the lazily-resolved framework back for the next test.
func resetLoad(t *testing.T) {
	t.Helper()
	open, bind := dsOpen, dsBind
	t.Cleanup(func() {
		dsOpen, dsBind = open, bind
		loadOnce, loadErr, getFn, setFn = sync.Once{}, nil, nil, nil
	})
	loadOnce, loadErr, getFn, setFn = sync.Once{}, nil, nil, nil
}

// TestAFrameworkThatWillNotOpenIsReportedOnce, not crashed into and not
// retried on every call: a system that has moved on has moved on.
func TestAFrameworkThatWillNotOpenIsReportedOnce(t *testing.T) {
	resetLoad(t)
	calls := 0
	dsOpen = func() (uintptr, error) { calls++; return 0, errors.New("gone") }

	if _, err := Of(1); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Of = %v, want ErrUnavailable", err)
	}
	if err := Set(1, Full); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Set = %v, want ErrUnavailable", err)
	}
	if _, err := Dim(1); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Dim = %v, want ErrUnavailable", err)
	}
	if calls != 1 {
		t.Errorf("the framework was looked for %d times, want 1", calls)
	}
}

// TestADisplayThatWillNotSayIsNotDimmed: there would be nothing to restore, and
// a zero read as "off" would darken a panel this package could never light
// again.
func TestADisplayThatWillNotSayIsNotDimmed(t *testing.T) {
	resetLoad(t)
	var sets int
	dsOpen = func() (uintptr, error) { return 1, nil }
	dsBind = func(uintptr) {
		getFn = func(uint32, *float32) int32 { return -1 } // refuses
		setFn = func(uint32, float32) int32 { sets++; return 0 }
	}

	if _, err := Dim(7); !errors.Is(err, ErrNoBrightness) {
		t.Fatalf("Dim = %v, want ErrNoBrightness", err)
	}
	if sets != 0 {
		t.Errorf("the display was set %d times after refusing to be read", sets)
	}
}

// TestDimHandsBackTheLevelItFound, which is the whole point: a program that
// darkens a screen and then dies leaves somebody with a black panel.
func TestDimHandsBackTheLevelItFound(t *testing.T) {
	resetLoad(t)
	level := float32(0.62)
	dsOpen = func() (uintptr, error) { return 1, nil }
	dsBind = func(uintptr) {
		getFn = func(_ uint32, out *float32) int32 { *out = level; return 0 }
		setFn = func(_ uint32, v float32) int32 { level = v; return 0 }
	}

	restore, err := Dim(1)
	if err != nil {
		t.Fatalf("Dim: %v", err)
	}
	if level != 0 {
		t.Errorf("the panel is at %v after Dim, want off", level)
	}
	if err := restore(); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if level != 0.62 {
		t.Errorf("restored to %v, want the 0.62 it found", level)
	}
}

// TestADisplayThatWillNotBeToldIsReported.
func TestADisplayThatWillNotBeToldIsReported(t *testing.T) {
	resetLoad(t)
	dsOpen = func() (uintptr, error) { return 1, nil }
	dsBind = func(uintptr) {
		getFn = func(_ uint32, out *float32) int32 { *out = 0.5; return 0 }
		setFn = func(uint32, float32) int32 { return -1 }
	}
	if err := Set(1, Full); !errors.Is(err, ErrNoBrightness) {
		t.Errorf("Set = %v, want ErrNoBrightness", err)
	}
	if _, err := Dim(1); !errors.Is(err, ErrNoBrightness) {
		t.Errorf("Dim = %v, want ErrNoBrightness", err)
	}
}

// TestLiveTheBuiltInPanelAnswersAndRestores, on the machine running this.
func TestLiveTheBuiltInPanelAnswersAndRestores(t *testing.T) {
	if os.Getenv("BRIGHTNESS_LIVE") == "" {
		t.Skip("set BRIGHTNESS_LIVE=1 to run the test that changes a real panel")
	}
	const mainDisplay = 1
	was, err := Of(mainDisplay)
	if err != nil {
		t.Skipf("display %d does not report a brightness: %v", mainDisplay, err)
	}
	t.Logf("display %d is at %.2f", mainDisplay, was)

	// A small step rather than off: this is somebody's screen.
	target := was - 0.2
	if target < 0 {
		target = was + 0.2
	}
	if err := Set(mainDisplay, target); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := Of(mainDisplay)
	if err != nil {
		t.Fatalf("Of: %v", err)
	}
	if diff := got - target; diff > 0.02 || diff < -0.02 {
		t.Errorf("asked for %.2f, reads back %.2f", target, got)
	}
	if err := Set(mainDisplay, was); err != nil {
		t.Fatalf("restoring: %v", err)
	}
	back, _ := Of(mainDisplay)
	if diff := back - was; diff > 0.02 || diff < -0.02 {
		t.Errorf("restored to %.2f, want the %.2f it was", back, was)
	}
}
