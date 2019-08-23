package storage

import (
	"time"

	"github.com/Quard/timekeeper/internal/timerecord"
)

type Storage interface {
	Add(record *timerecord.TimeRecord) error
	Finish(userID string, id string) (*timerecord.TimeRecord, error)
	GetGroupedByTask(userID string, date time.Time) <-chan timerecord.GroupedReport
}
