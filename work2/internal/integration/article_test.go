package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"example/wb/internal/domain"
	"example/wb/internal/integration/startup"
	"example/wb/internal/repository/dao"
	ijwt "example/wb/internal/web/jwt"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ArticleMongoDBHandlerSuite struct {
	suite.Suite
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
	server  *gin.Engine
}

func (s *ArticleMongoDBHandlerSuite) SetupSuite() {
	s.mdb = startup.InitMongoDB()
	err := dao.InitCollections(s.mdb)
	assert.NoError(s.T(), err)
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	fmt.Println(node)

	hdl := startup.InitArticleHandler(dao.NewMongoDBArticleDAO(s.mdb, node))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		// 设置用户登录态
		ctx.Set("user", ijwt.UserClaims{
			Id: 123,
		})
		// ctx.Next()
	})
	hdl.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleMongoDBHandlerSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.liveCol.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoDBHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string
		// 提前准备数据的位置
		before func(t *testing.T)
		// 验证数据的位置
		after      func(t *testing.T)
		req        Article
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {
				// 什么都不需要做
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "author_id", Value: 123}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				art.Utime = 0
				art.Ctime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished,
				}, art)
			},
			req: Article{
				Title:   "测试用例1",
				Content: "这是我的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "更新帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    432,
					Utime:    4324,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "id", Value: 111}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Utime > 4324)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       111,
					Title:    "新的标题1",
					Content:  "新的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    432,
				}, art)
			},
			req: Article{
				Id:      111,
				Title:   "新的标题1",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 11,
			},
		},
		{
			name: "更新别人的帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       11,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 1234,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    432,
					Utime:    4324,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "id", Value: 11}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       11,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 1234,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    432,
					Utime:    4324,
				}, art)
			},
			req: Article{
				Id:      11,
				Title:   "新的标题1",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost,
				"/article/edit", bytes.NewBuffer(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, res.Code)
			if res.Code == 0 {
				assert.True(t, tc.wantResult.Data > 0)
			}
			tc.after(t)
		})
	}
}

func (s *ArticleMongoDBHandlerSuite) TestPublished() {
	t := s.T()
	testCases := []struct {
		name string
		// 提前准备数据的位置
		before func(t *testing.T)
		// 验证数据的位置
		after      func(t *testing.T)
		req        Article
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子, 并发表",
			before: func(t *testing.T) {
				// 什么都不需要做
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "author_id", Value: 123}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				var liveArt dao.Article
				err = s.liveCol.FindOne(ctx, filter).Decode(&liveArt)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Id == liveArt.Id)
				assert.True(t, art.Title == liveArt.Title)
				assert.True(t, art.Content == liveArt.Content)
				assert.True(t, art.AuthorId == liveArt.AuthorId)
				assert.True(t, int64(art.Status) == int64(liveArt.Status))
				assert.True(t, art.Ctime == liveArt.Ctime)
				assert.True(t, art.Utime == liveArt.Utime)
				assert.True(t, art == liveArt)
				art.Id = 0
				art.Utime = 0
				art.Ctime = 0
				assert.Equal(t, dao.Article{
					Title:    "新建一个用于发表的帖子",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished,
				}, art)
			},
			req: Article{
				Title:   "新建一个用于发表的帖子",
				Content: "这是我的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			// 线上表没有，但是制作表有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    432,
					Utime:    4324,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "author_id", Value: 123},
					{Key: "id", Value: 111}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				var liveArt dao.Article
				err = s.liveCol.FindOne(ctx, filter).Decode(&liveArt)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Status == domain.ArticleStatusPublished)
				assert.True(t, art.Id == liveArt.Id)
				assert.True(t, art.Title == liveArt.Title)
				assert.True(t, art.Content == liveArt.Content)
				assert.True(t, art.AuthorId == liveArt.AuthorId)
				assert.True(t, int64(art.Status) == int64(liveArt.Status))
				assert.True(t, art.Ctime < liveArt.Ctime)
				assert.True(t, art.Utime == liveArt.Utime)
				art.Id = 0
				art.Utime = 0
				art.Ctime = 0
				assert.Equal(t, dao.Article{
					Title:    "更新一个用于发表的帖子",
					Content:  "这是我新的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished,
				}, art)
			},
			req: Article{
				Id:      111,
				Title:   "更新一个用于发表的帖子",
				Content: "这是我新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			// 线上表有，制作表也有
			name: "更新帖子, 并重新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       1111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    432,
					Utime:    4324,
				})
				assert.NoError(t, err)
				_, err = s.liveCol.InsertOne(ctx, dao.Article{
					Id:       1111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    439,
					Utime:    4324,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "author_id", Value: 123},
					{Key: "id", Value: 1111}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				var liveArt dao.Article
				err = s.liveCol.FindOne(ctx, filter).Decode(&liveArt)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Status == domain.ArticleStatusPublished)
				assert.True(t, art.Id == liveArt.Id)
				assert.True(t, art.Title == liveArt.Title)
				assert.True(t, art.Content == liveArt.Content)
				assert.True(t, art.AuthorId == liveArt.AuthorId)
				assert.True(t, int64(art.Status) == int64(liveArt.Status))
				assert.True(t, art.Ctime < liveArt.Ctime)
				assert.True(t, art.Utime == liveArt.Utime)
				art.Id = 0
				art.Utime = 0
				art.Ctime = 0
				assert.Equal(t, dao.Article{
					Title:    "更新一个用于发表的帖子",
					Content:  "这是我新的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished,
				}, art)
			},
			req: Article{
				Id:      1111,
				Title:   "更新一个用于发表的帖子",
				Content: "这是我新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			// 线上表有，制作表也有
			name: "更新别人的帖子, 发表失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       11111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 1234,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    432,
					Utime:    4324,
				})
				assert.NoError(t, err)
				_, err = s.liveCol.InsertOne(ctx, dao.Article{
					Id:       11111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 1234,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    439,
					Utime:    4324,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "author_id", Value: 1234},
					{Key: "id", Value: 11111}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				var liveArt dao.Article
				err = s.liveCol.FindOne(ctx, filter).Decode(&liveArt)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Id == liveArt.Id)
				assert.True(t, art.Title == liveArt.Title)
				assert.True(t, art.Content == liveArt.Content)
				assert.True(t, art.AuthorId == liveArt.AuthorId)
				assert.True(t, int64(art.Status) == int64(liveArt.Status))
				assert.True(t, art.Ctime < liveArt.Ctime)
				assert.True(t, art.Utime == liveArt.Utime)
				art.Utime = 0
				art.Ctime = 0
				assert.Equal(t, dao.Article{
					Id:       11111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 1234,
					Status:   domain.ArticleStatusUnpublished,
				}, art)
			},
			req: Article{
				Id:      11111,
				Title:   "更新一个用于发表的帖子",
				Content: "这是我新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost,
				"/article/publish", bytes.NewBuffer(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, res.Code)
			if res.Code == 0 {
				assert.True(t, tc.wantResult.Data > 0)
			}
			tc.after(t)
		})
	}
}
func (s *ArticleMongoDBHandlerSuite) TestWithdraw() {
	t := s.T()
	testCases := []struct {
		name string
		// 提前准备数据的位置
		before func(t *testing.T)
		// 验证数据的位置
		after      func(t *testing.T)
		req        Article
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "撤回已经发布的文章",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       1111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished,
					Ctime:    432,
					Utime:    4324,
				})
				assert.NoError(t, err)
				_, err = s.liveCol.InsertOne(ctx, dao.Article{
					Id:       1111,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished,
					Ctime:    439,
					Utime:    4324,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "author_id", Value: 123},
					{Key: "id", Value: 1111}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				var liveArt dao.Article
				err = s.liveCol.FindOne(ctx, filter).Decode(&liveArt)
				assert.NoError(t, err)
				assert.True(t, art.Status == domain.ArticleStatusPrivate)
				assert.True(t, liveArt.Status == domain.ArticleStatusPrivate)
			},
			req: Article{
				Id: 1111,
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Msg: "OK",
			},
		},
		{
			name: "撤回别人的文章, 失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       11,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 1233,
					Status:   domain.ArticleStatusPublished,
					Ctime:    432,
					Utime:    4324,
				})
				assert.NoError(t, err)
				_, err = s.liveCol.InsertOne(ctx, dao.Article{
					Id:       11,
					Title:    "测试用例1",
					Content:  "这是我的内容",
					AuthorId: 1233,
					Status:   domain.ArticleStatusPublished,
					Ctime:    439,
					Utime:    4324,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				filter := bson.D{{Key: "author_id", Value: 1233},
					{Key: "id", Value: 11}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				var liveArt dao.Article
				err = s.liveCol.FindOne(ctx, filter).Decode(&liveArt)
				assert.NoError(t, err)
				assert.True(t, art.Status == domain.ArticleStatusPublished)
				assert.True(t, liveArt.Status == domain.ArticleStatusPublished)
			},
			req: Article{
				Id: 11,
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost,
				"/article/withdraw", bytes.NewBuffer(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, res.Code)
			if res.Code == 0 {
				assert.True(t, tc.wantResult.Msg == "OK")
			}
			if res.Code == 5 {
				assert.True(t, tc.wantResult.Msg == "系统错误")
			}
			tc.after(t)
		})
	}
}

func TestMongoArticle(t *testing.T) {
	suite.Run(t, new(ArticleMongoDBHandlerSuite))
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int64
	Msg  string
	Data T
}
