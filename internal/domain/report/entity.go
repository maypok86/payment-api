package report

import "errors"

var (
	ErrIsNotAvailable = errors.New("this report is not available yet")
	ErrNotFound       = errors.New("report not found")
)
