// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build !darwin

package brightness

// The seams answer that they cannot, rather than being absent: a consumer that
// cross-compiles gets the same API and one clean error from it.
func get(uint32) (float64, error) { return 0, ErrUnsupported }
func set(uint32, float64) error   { return ErrUnsupported }
