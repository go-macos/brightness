# go-macos/brightness

[![ci](https://github.com/go-macos/brightness/actions/workflows/ci.yml/badge.svg)](https://github.com/go-macos/brightness/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-macos/brightness.svg)](https://pkg.go.dev/github.com/go-macos/brightness)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)

**Read and set a display's backlight on macOS, from pure Go with
`CGO_ENABLED=0`.** No cgo, no `osascript`, no shelling out to anything.

```go
restore, err := brightness.Dim(display)
if err == nil {
        defer restore()      // ← the whole point
}
```

## Why

For the case where a screen is being shown **somewhere else**. A pair of display
glasses can mirror the Mac's own screen, and then the physical panel is worse
than useless: it is a copy of private work, at reading distance, facing whoever
walks past — and it is lit at full power for nobody.

Turning the backlight off beats covering the screen with a black window:

| | |
|---|---|
| a capture of that display is **unchanged** | the framebuffer is untouched, so a mirror keeps working |
| **nothing to exclude** | no window id to keep out of the capture, no stream to rebuild |
| **nothing can get above it** | a window can be covered by whatever raises itself next |
| the panel is **off** | not black at full power |

## The restore is the API

A program that darkens a screen and then exits — or crashes, or is killed —
leaves somebody with a black panel and no idea why. So `Dim` reads the level
**before** changing anything and hands it straight back as the way home. Keep it
and defer it.

A display that will not say what its brightness is, is not dimmed: there would
be nothing to restore, and a zero read as "off" would darken a panel this
package could never light again.

## ⚠ This is PRIVATE API

`DisplayServices` is a private framework. Apple may change or remove it in any
release, and a program linking this cannot ship on the Mac App Store.

The documented route — `IODisplaySetFloatParameter` with
`kIODisplayBrightnessKey` — is the **old** one: it does not drive the built-in
panel on an Apple Silicon Mac. This is what does.

Every symbol is looked up at run time and a failure is reported rather than
crashing, so a system that has moved on says so once (`ErrUnavailable`) instead
of being asked again on every call.

## What answers

The built-in panel answers. An external display answers only if it speaks DDC/CI
and macOS is willing to use it — many do not, and `Of` reports `ErrNoBrightness`
rather than pretending.

## Tests

```
go test ./...                          # portable logic, and every failure path
BRIGHTNESS_LIVE=1 go test -run Live    # + a real panel, moved and put back
```

The failure paths are driven through seams — the framework that will not open,
the display that will not say, the display that will not be told — so they are
exercised on a machine where all of it works. The live test moves the real panel
by a fifth and puts it back, because it is somebody's screen.

## Requirements

macOS, Go 1.24+, `CGO_ENABLED=0`. Off darwin every entry point answers
`ErrUnsupported`.

## Licence

BSD-3-Clause. See [LICENSE](LICENSE).
