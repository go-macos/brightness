// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

// Package brightness reads and sets a display's backlight on macOS, from pure
// Go with CGO_ENABLED=0.
//
// It exists for the case where a screen is being shown SOMEWHERE ELSE. A pair
// of display glasses can mirror the Mac's own screen, and then the physical
// panel is worse than useless: it is a copy of private work, at reading
// distance, facing whoever walks past — and it is lit at full power for nobody.
//
//	restore, err := brightness.Dim(display)
//	if err == nil {
//		defer restore()
//	}
//
// Turning the backlight off is better than covering the screen with a black
// window: nothing changes in a capture of that display (the framebuffer is
// untouched), there is no window for a capture to exclude and no stream to
// rebuild, and no window that something else can raise itself above.
//
// # ⚠ This is PRIVATE API
//
// DisplayServices is a private framework. Apple may change or remove it in any
// release; a program linking this cannot ship on the Mac App Store. Every
// symbol is looked up at run time and a failure is reported rather than
// crashing, so a system that has moved on says so.
//
// # What answers
//
// The built-in panel answers. An external display answers only if it speaks
// DDC/CI and macOS is willing to use it — many do not, and [Of] then reports
// [ErrNoBrightness] rather than pretending.
package brightness

import "errors"

// Errors this package returns.
var (
	// ErrUnsupported is what every entry point answers away from macOS.
	ErrUnsupported = errors.New("brightness: unsupported on this platform (macOS only)")
	// ErrUnavailable means the private framework is not there, or no longer
	// exports what this package needs.
	ErrUnavailable = errors.New("brightness: the DisplayServices framework is not available")
	// ErrNoBrightness means that display will not say, or will not be told —
	// an external panel with no DDC, most often.
	ErrNoBrightness = errors.New("brightness: this display does not report a brightness")
	// ErrRange means a level outside 0..1 was asked for.
	ErrRange = errors.New("brightness: a level must be between 0 and 1")
)

// Off is a dark panel, and Full is a bright one.
const (
	Off  = 0.0
	Full = 1.0
)

// Of returns the display's brightness, from 0 (dark) to 1 (full).
func Of(display uint32) (float64, error) { return get(display) }

// Set puts the display's brightness at level, which must be between [Off] and
// [Full].
//
// It is not a fade: the panel arrives at once, which is what a person pressing
// a key expects and what a program restoring what it changed needs.
func Set(display uint32, level float64) error {
	if level < Off || level > Full {
		return ErrRange
	}
	return set(display, level)
}

// Dim turns the display off and returns the way back.
//
// The restore is the point of the whole package. A program that darkens a
// screen and then exits — or crashes, or is killed — leaves somebody with a
// black panel and no idea why, so the value to go back to is read BEFORE
// anything is changed and handed straight back to the caller. Keep it and defer
// it.
//
// A display that will not say what its brightness is, is not dimmed: there
// would be nothing to restore.
func Dim(display uint32) (restore func() error, err error) {
	was, err := Of(display)
	if err != nil {
		return nil, err
	}
	if err := Set(display, Off); err != nil {
		return nil, err
	}
	return func() error { return Set(display, was) }, nil
}
