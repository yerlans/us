package storage

import (
	"fmt"
)

var (
	ErrURLNotFound = fmt.Errorf("url not found")
	ErrURLExists   = fmt.Errorf("url already exists")
)
