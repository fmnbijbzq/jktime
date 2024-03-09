package mongodb

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCrud(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
	assert.NoError(t, err)
	mdb := client.Database("webook")
	col := mdb.Collection("articles")
	defer func() {
		_, err = col.DeleteMany(ctx, bson.D{})
	}()

	res, err := col.InsertMany(ctx, []interface{}{
		Article{Id: 1, Title: "第一个11", Content: "32432432sdhfaojS附近的撒"},
		Article{Id: 2, Title: "第二个22", Content: "地方撒绘画覅近的撒"},
	})
	assert.NoError(t, err)
	fmt.Printf("---------id %s \n", res.InsertedIDs...)

	// bson
	// 找 ID = 2 的
	filter := bson.D{bson.E{Key: "id", Value: 2}}
	var art Article
	err = col.FindOne(ctx, filter).Decode(&art)
	if err == mongo.ErrNoDocuments {
		fmt.Println("没有数据")
	}
	assert.NoError(t, err)
	fmt.Printf("%#v \n", art)
	fmt.Printf("%v \n", art)
	// 找不存在
	filter = bson.D{bson.E{Key: "id", Value: 4}}
	err = col.FindOne(ctx, filter).Decode(&art)
	if err == mongo.ErrNoDocuments {
		fmt.Println("没有数据")
	}
	fmt.Printf("%#v \n", art)
	fmt.Printf("%v \n", art)
	// 更新id=2的数据
	filter = bson.D{{Key: "id", Value: 2}}
	updater := bson.D{{"$set", bson.D{{"content", "修改成"}, {"author_id", 1}}}}
	ures, err := col.UpdateOne(ctx, filter, updater)
	assert.NoError(t, err)
	fmt.Printf("--%#v\n", ures)

	filter = bson.D{bson.E{Key: "id", Value: 2}}
	err = col.FindOne(ctx, filter).Decode(&art)
	if err == mongo.ErrNoDocuments {
		fmt.Println("没有数据")
	}
	assert.NoError(t, err)
	fmt.Printf("%#v \n", art)
	fmt.Printf("%v \n", art)

	// or
	or := bson.A{bson.D{{"id", 1}}, bson.D{{"id", 3}}}
	orRes, err := col.Find(ctx, bson.D{bson.E{"$or", or}})
	assert.NoError(t, err)
	var ars []Article
	err = orRes.All(ctx, &ars)
	assert.NoError(t, err)
	fmt.Printf("Or---%#v \n", ars)

	// and
	and := bson.A{bson.D{{Key: "id", Value: 1}},
		bson.D{{Key: "title", Value: "第一个11"}},
	}
	andRes, err := col.Find(ctx, bson.D{bson.E{"$and", and}})
	assert.NoError(t, err)
	err = andRes.All(ctx, &ars)
	assert.NoError(t, err)
	fmt.Printf("And---%#v \n", ars)

	// in
	in := bson.A{bson.D{bson.E{"content", bson.M{"$in": []any{"修改成功"}}}},
		bson.D{bson.E{"id", bson.M{"$in": []any{2}}}}}
	// in := bson.D{bson.E{"id", bson.M{"$in": []any{123, 456}}}}
	inRes, err := col.Find(ctx, bson.D{bson.E{"$or", in}},
		options.Find().SetProjection(bson.M{
			"title": "第一个11",
		}))
	assert.NoError(t, err)
	err = inRes.All(ctx, &ars)
	assert.NoError(t, err)
	fmt.Printf("In---%#v \n", ars)

	idxRes, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"author_id": 1},
		},
	})
	assert.NoError(t, err)
	fmt.Printf("Index----%#v\n", idxRes)

}

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"`
	AuthorId int64  `bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}
