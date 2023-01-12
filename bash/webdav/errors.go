package webdav

import (
	"errors"
	"fmt"
)

var (
	ErrNeedPrint           = errors.New("")
	ErrEditTargetNotExists = fmt.Errorf("the edit target is not exists. %w", ErrNeedPrint)
)
