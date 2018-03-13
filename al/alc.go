// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux windows

package al

import (
	"errors"
	"sync"
	"unsafe"
	"fmt"
)

var (
	mu      sync.Mutex
	device  unsafe.Pointer
	context unsafe.Pointer
)

// DeviceError returns the last known error from the current device.
func DeviceError() int32 {
	return alcGetError(device)
}

// TODO(jbd): Investigate the cases where multiple audio output
// devices might be needed.

// OpenDevice opens the default audio device.
// Calls to OpenDevice are safe for concurrent use.
func OpenDevice() error {
	mu.Lock()
	defer mu.Unlock()

	// already opened
	if device != nil {
		return nil
	}

	dev := alcOpenDevice("")
	if dev == nil {
		return errors.New("al: cannot open the default audio device")
	}
	ctx := alcCreateContext(dev, nil)
	if ctx == nil {
		alcCloseDevice(dev)
		return errors.New("al: cannot create a new context")
	}
	if !alcMakeContextCurrent(ctx) {
		alcCloseDevice(dev)
		return errors.New("al: cannot make context current")
	}
	device = dev
	context = ctx
	return nil
}

// CloseDevice closes the device and frees related resources.
// Calls to CloseDevice are safe for concurrent use.
func CloseDevice() {
	mu.Lock()
	defer mu.Unlock()

	if device == nil {
		return
	}

	alcCloseDevice(device)
	if context != nil {
		alcDestroyContext(context)
	}
	device = nil
	context = nil
}

func GetError() error {
	if device == nil {
		return errors.New("OpenAL error: Divice isn't open")
	}
	c := alcGetError(device)
	switch c {
	case 0:
		return nil
	case InvalidName:
		return errors.New("OpenAL error: invalid device")
	case InvalidContext:
		return errors.New("OpenAL error: invalid context")
	case InvalidEnum:
		return errors.New("OpenAL error: invalid enum")
	case InvalidValue:
		return errors.New("OpenAL error: invalid value")
	case OutOfMemory:
		return errors.New("OpenAL error: out of memory")
	default:
		return fmt.Errorf("OpenAL error: code %d", c)
	}
}
