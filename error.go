// Copyright (c) 2021 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"errors"
)

// ErrClosed is the error used for read/write operations on a closed io.
var ErrClosed = errors.New("speedio: closed")

// ErrZeroWrite is the error thrown when zero bytes written.
var ErrZeroWrite = errors.New("speedio: zero bytes written")

// ErrInvalidParameter is the error thrown when a parameter is invalid.
var ErrInvalidParameter = errors.New("speedio: invalid parameter")
