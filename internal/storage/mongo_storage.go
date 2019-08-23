package storage

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Quard/timekeeper/internal/timerecord"
)

const timeCollectionName = "time"

type MongoStorage struct {
	uri  string
	conn *mongo.Database
}

func NewMongoStorage(uri string) MongoStorage {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	return MongoStorage{
		uri:  uri,
		conn: client.Database("timekeeper"),
	}
}

func (s MongoStorage) Add(record *timerecord.TimeRecord) error {
	timeCollection := s.conn.Collection(timeCollectionName)
	s.DoneExistingRecords(record.UserID)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	doc, errConv := asNewDocument(record)
	if errConv != nil {
		return errConv
	}
	newRecord, err := timeCollection.InsertOne(ctx, doc)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	record.ID = newRecord.InsertedID.(primitive.ObjectID)

	return nil
}

func (s MongoStorage) Finish(userID string, id string) (*timerecord.TimeRecord, error) {
	userObjectID, errConvUserID := primitive.ObjectIDFromHex(userID)
	if errConvUserID != nil {
		sentry.CaptureException(errConvUserID)
		return nil, errors.New("auth error")
	}

	getFilter := bson.M{"user_id": userObjectID}
	updateFilter := bson.M{"user_id": userObjectID, "time_end": time.Time{}}
	if id != "" {
		objID, errConvObjID := primitive.ObjectIDFromHex(id)
		if errConvObjID != nil {
			sentry.CaptureException(errConvObjID)
			return nil, errors.New("auth error")
		}
		updateFilter = bson.M{"user_id": userID, "time_end": time.Time{}, "_id": objID}
		getFilter = bson.M{"user_id": userID, "_id": objID}
	}

	timeCollection := s.conn.Collection(timeCollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := timeCollection.UpdateMany(
		ctx,
		updateFilter,
		bson.M{"$set": bson.M{"time_end": time.Now()}},
	)
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	record, err := s.getRecordByFilter(getFilter)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s MongoStorage) GetGroupedByTask(userID string, date time.Time) <-chan timerecord.GroupedReport {
	userObjectID, errConvUserID := primitive.ObjectIDFromHex(userID)
	if errConvUserID != nil {
		sentry.CaptureException(errConvUserID)
		return nil
	}

	ch := make(chan timerecord.GroupedReport)

	go func() {
		defer close(ch)

		timeCollection := s.conn.Collection(timeCollectionName)
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		opts := &options.FindOptions{}
		opts.SetSort(bson.M{"task": 1})
		filter := bson.M{
			"user_id": userObjectID,
			"time_start": bson.M{
				"$gte": date,
				"$lt":  date.Add(24 * time.Hour),
			},
		}
		log.Println(filter)
		cursor, err := timeCollection.Find(ctx, filter, opts)
		if err != nil {
			sentry.CaptureException(err)
			return
		}
		defer cursor.Close(context.Background())

		report := timerecord.GroupedReport{Comments: []string{}}
		for cursor.Next(ctx) {
			var record timerecord.TimeRecord
			if err := cursor.Decode(&record); err != nil {
				sentry.CaptureException(err)
				continue
			}
			log.Println("-->", record)
			if strings.Compare(record.Task, report.Task) != 0 {
				if report.SpentTime > 0 {
					log.Println("==>", report)
					ch <- report
				}

				report = timerecord.GroupedReport{
					Task:      record.Task,
					SpentTime: 0,
					Comments:  []string{},
				}
			}

			duration := record.TimeEnd.Sub(record.TimeStart)
			log.Println(duration)
			report.SpentTime = report.SpentTime + int64(duration.Minutes()) + int64(duration.Hours())*60

			if record.Comment != "" {
				report.Comments = append(report.Comments, record.Comment)
			}
		}

		if report.SpentTime > 0 {
			ch <- report
		}

	}()

	return ch
}

func (s MongoStorage) DoneExistingRecords(userID primitive.ObjectID) error {
	timeCollection := s.conn.Collection(timeCollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := timeCollection.UpdateMany(
		ctx,
		bson.M{"user_id": userID, "time_end": time.Time{}},
		bson.M{"$set": bson.M{"time_end": time.Now()}},
	)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	return nil
}

func (s MongoStorage) getRecordByFilter(filter bson.M) (*timerecord.TimeRecord, error) {
	timeCollection := s.conn.Collection(timeCollectionName)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	opts := &options.FindOneOptions{}
	opts.SetSort(bson.M{"time_start": -1})
	result := timeCollection.FindOne(ctx, filter, opts)
	if result.Err() != nil {
		sentry.CaptureException(result.Err())
		return nil, result.Err()
	}

	var record timerecord.TimeRecord
	if err := result.Decode(&record); err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	return &record, nil
}

func asNewDocument(val interface{}) (bson.M, error) {
	var document bson.M
	bytes, err := bson.Marshal(val)
	if err != nil {
		sentry.CaptureException(err)
		return document, err
	}
	if err := bson.Unmarshal(bytes, &document); err != nil {
		sentry.CaptureException(err)
		return document, err
	}

	delete(document, "_id")

	return document, nil
}
