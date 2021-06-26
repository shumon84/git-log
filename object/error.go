package object

import "errors"

var (
	ErrInvalidObject       = errors.New("invalid object")
	ErrNotCommitObject     = errors.New("not commit object")
	ErrInvalidCommitObject = errors.New("invalid commit object")
)
