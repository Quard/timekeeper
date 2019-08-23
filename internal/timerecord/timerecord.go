package timerecord

import (
	"time"

	"github.com/getsentry/sentry-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TimeRecord struct {
	ID        primitive.ObjectID `bson:"_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Task      string             `bson:"task"`
	TimeStart time.Time          `bson:"time_start"`
	TimeEnd   time.Time          `bson:"time_end"`
	Comment   string             `bson:"comment"`
}

type GroupedReport struct {
	Task      string
	SpentTime int64
	Comments  []string
}

func NewTimeRecord(userID string, task string, timeStart time.Time, comment string) (TimeRecord, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		sentry.CaptureException(err)
		return TimeRecord{}, err
	}

	timeRecord := TimeRecord{
		UserID:    userObjectID,
		Task:      task,
		TimeStart: timeStart,
		Comment:   comment,
	}

	return timeRecord, nil
}
