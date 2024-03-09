package dao

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

// GetByAuthor implements ArticleDAO.
func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	panic("unimplemented")
}

// GetById implements ArticleDAO.
func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	panic("unimplemented")
}

// GetPubById implements ArticleDAO.
func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	panic("unimplemented")
}

// Insert implements ArticleDao.
func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	art.Id = m.node.Generate().Int64()
	now := time.Now().UnixMilli()
	utime := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = utime
	_, err := m.col.InsertOne(ctx, &art)
	return art.Id, err
}

// Sync implements ArticleDao.
func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	now := time.Now().UnixMilli()
	art.Utime = now
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	filter := bson.D{
		{Key: "id", Value: id},
		{Key: "author_id", Value: art.AuthorId},
	}
	set := bson.D{{Key: "$set", Value: art},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "ctime", Value: now},
		}}}
	_, err = m.liveCol.UpdateOne(ctx, filter, set,
		options.Update().SetUpsert(true))
	return id, err
}

// SyncStatus implements ArticleDao.
func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	filter := bson.D{
		{Key: "id", Value: id},
		{Key: "author_id", Value: uid},
	}
	set := bson.D{{Key: "$set",
		Value: bson.D{bson.E{Key: "status", Value: status}}},
	}
	// Value: bson.M{"status": status}},
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("ID 不对或者创作者不对")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, set)
	return err
}

// UpdateById implements ArticleDao.
func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{{Key: "$set", Value: bson.M{
		"title":   art.Title,
		"content": art.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  art.Status,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBArticleDAO{
		node:    node,
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
	}
}
