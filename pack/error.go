package pack

import "errors"

var (
	ErrNotFoundIdxEntry        = errors.New("not found idx entry")
	ErrCantBeConvertedToObject = errors.New("can't be converted to object")
	ErrInvalidPackFile         = errors.New("invalid pack file")
)
