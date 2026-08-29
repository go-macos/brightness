// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin

package brightness

import (
	"sync"

	"github.com/ebitengine/purego"
)

// displayServices is the private framework that owns the backlight on Apple
// Silicon.
//
// IODisplaySetFloatParameter with kIODisplayBrightnessKey is the documented
// route and it is the OLD one: it does not drive the built-in panel on an
// Apple Silicon Mac. This is what does, and it is private — see the package
// comment.
const displayServices = "/System/Library/PrivateFrameworks/DisplayServices.framework/DisplayServices"

// The seams. They are package vars so a test can drive every branch — the
// framework that will not open, the display that will not answer — on a machine
// where both work.
var (
	loadOnce sync.Once
	loadErr  error

	getFn func(display uint32, out *float32) int32
	setFn func(display uint32, level float32) int32

	dsOpen = func() (uintptr, error) {
		return purego.Dlopen(displayServices, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	}
	dsBind = func(h uintptr) {
		purego.RegisterLibFunc(&getFn, h, "DisplayServicesGetBrightness")
		purego.RegisterLibFunc(&setFn, h, "DisplayServicesSetBrightness")
	}
)

// load resolves the framework once.
func load() error {
	loadOnce.Do(func() {
		h, err := dsOpen()
		if err != nil {
			loadErr = ErrUnavailable
			return
		}
		dsBind(h)
	})
	return loadErr
}

func platformGet(display uint32) (float64, error) {
	if err := load(); err != nil {
		return 0, err
	}
	var v float32
	// A NON-ZERO status is a display that will not say. It is the ordinary
	// answer for an external panel with no DDC, so it is an error a caller can
	// test for rather than a zero brightness, which would read as "off".
	if st := getFn(display, &v); st != 0 {
		return 0, ErrNoBrightness
	}
	return float64(v), nil
}

func platformSet(display uint32, level float64) error {
	if err := load(); err != nil {
		return err
	}
	if st := setFn(display, float32(level)); st != 0 {
		return ErrNoBrightness
	}
	return nil
}
