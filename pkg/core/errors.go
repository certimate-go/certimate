package core

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented = fmt.Errorf("not implemented function")
	ErrUnsupported    = errors.ErrUnsupported
)
