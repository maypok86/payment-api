package report

import (
	"errors"
	"time"
)

var (
	ErrAlreadyExist = errors.New("report with given fields already exist")
	ErrNotFound     = errors.New("report not found")
)

type Report struct {
	ReportID  int64
	ServiceID int64
	Amount    int64
	CreatedAt time.Time
}
