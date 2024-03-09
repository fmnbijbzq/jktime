package startup

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() *mongo.Database {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			log.Printf("mongo reqId:%d start on db:%s cmd:%s sql:%+v",
				cse.RequestID, cse.DatabaseName,
				cse.CommandName, cse.Command,
			)
		},
		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
			log.Printf("mongo reqId:%d exec cmd:%s success duration %d ns", succeededEvent.RequestID,
				succeededEvent.CommandName, succeededEvent.DurationNanos)
		},
		Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
			log.Printf("mongo reqId:%d exec cmd:%s failed duration %d ns", failedEvent.RequestID,
				failedEvent.CommandName, failedEvent.DurationNanos)
		},
	}
	clientOpt := options.Client().
		ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, clientOpt)
	if err != nil {
		panic(err)
	}
	return client.Database("webook")

}
