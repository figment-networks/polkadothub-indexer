package types

import (
	"fmt"
	"github.com/pkg/errors"
	"time"
)

const (
	IntervalHourly SummaryInterval = "hour"
	IntervalDaily  SummaryInterval = "day"
)

// SummaryInterval type represents summary interval
type SummaryInterval string

func (s SummaryInterval) Valid() bool {
	return s == IntervalHourly || s == IntervalDaily
}

func (s SummaryInterval) Equal(o SummaryInterval) bool {
	return s == o
}

func (s SummaryInterval) Duration() (*time.Duration, error) {
	var durationInterval string
	if s == IntervalHourly {
		durationInterval = "1h"
	} else if s == IntervalDaily {
		durationInterval = "24h"
	} else {
		return nil, errors.New(fmt.Sprintf("unknown summary interval %s", s))
	}
	duration, err := time.ParseDuration(durationInterval)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}