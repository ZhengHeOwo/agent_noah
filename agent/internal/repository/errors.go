package repository

import (
	"errors"
	"fmt"
)

var (
	ErrTransient = errors.New("Transient failure")

	ErrCreate = fmt.Errorf("Resource creation failed")

	ErrGet = errors.New("Resouce acquisition failed")
)
