package timekeeper_rpc

import (
	context "context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Quard/timekeeper/internal/timerecord"
)

func (srv TimeKeeperRPC) Add(ctx context.Context, rec *TimeStartRecord) (*TimeRecord, error) {
	timeStart := time.Unix(rec.TimeStart, 0)
	record, err := timerecord.NewTimeRecord(rec.UserID, rec.Task, timeStart, rec.Comment)
	if err != nil {
		return nil, err
	}

	err = srv.storage.Add(&record)
	if err != nil {
		return nil, err
	}

	newRecord := TimeRecord{
		ID:        record.ID.Hex(),
		UserID:    record.UserID.Hex(),
		Task:      record.Task,
		TimeStart: record.TimeStart.Unix(),
		Comment:   record.Comment,
	}

	return &newRecord, nil
}

func (srv TimeKeeperRPC) Done(ctx context.Context, rec *RecordID) (*TimeRecord, error) {
	record, err := srv.storage.Finish(rec.UserID, rec.ID)
	if err != nil {
		return nil, errors.New("unable to finish currect task")
	}

	newRecord := TimeRecord{
		ID:        record.ID.Hex(),
		UserID:    record.UserID.Hex(),
		Task:      record.Task,
		TimeStart: record.TimeStart.Unix(),
		TimeEnd:   record.TimeEnd.Unix(),
		Comment:   record.Comment,
	}

	return &newRecord, nil
}

func (srv TimeKeeperRPC) GetGroupedForDate(r *GetForDate, s TimeKeeperRPC_GetGroupedForDateServer) error {
	datetime := time.Unix(r.Date, 0)
	date := time.Date(datetime.Year(), datetime.Month(), datetime.Day(), 0, 0, 0, 0, time.UTC)

	ch := srv.storage.GetGroupedByTask(r.UserID, date)
	for record := range ch {
		log.Println("~~>", record)
		s.Send(
			&TimeGroupedRecord{
				UserID:       r.UserID,
				Task:         record.Task,
				Date:         date.Unix(),
				SpentMinutes: record.SpentTime,
				Comments:     strings.Join(record.Comments, "; "),
			},
		)

	}

	return nil
}
