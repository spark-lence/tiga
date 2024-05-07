package tiga

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongodbDao struct {
	db     *mongo.Database
	config *Configuration
}
type MongoStruct interface {
	Save() error
	Read() interface{}
}

func NewMongodbDao(config *Configuration) *MongodbDao {
	env := config.GetEnv()
	username := config.GetConfigByEnv(env, "mongodb.username")
	password := config.GetConfigByEnv(env, "mongodb.password")
	host := config.GetConfigByEnv(env, "mongodb.host")

	port := config.GetConfigByEnv(env, "mongodb.port")
	db := config.GetConfigByEnv(env, "mongodb.db").(string)
	size := config.GetConfigByEnv(env, "mongodb.connectSize").(int)
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?connect=direct&&authSource=dev", username, password, host, port, db)

	// 设置连接超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	// 通过传进来的uri连接相关的配置
	o := options.Client().ApplyURI(uri)
	// 设置最大连接数 - 默认是100 ，不设置就是最大 max 64
	o.SetMaxPoolSize(uint64(size))
	// 发起链接
	client, err := mongo.Connect(ctx, o)
	if err != nil {
		panic(err)
	}
	// 判断服务是不是可用
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	// 返回 client
	return &MongodbDao{
		db:     client.Database(db),
		config: config,
	}

}
func NewMongoMocker(db *mongo.Database, config *Configuration) *MongodbDao {
	return &MongodbDao{
		db:     db,
		config: config,
	}

}
func (m MongodbDao) Upsert(ctx context.Context, collection string, filter interface{}, update primitive.M) error {
	c := m.db.Collection(collection)
	opts := options.Update().SetUpsert(true)
	_, err := c.UpdateOne(ctx, filter, update, opts)
	if err != nil && !strings.Contains(err.Error(), "no documents in result") {
		return err
	}
	return nil
}
func (m MongodbDao) Insert(ctx context.Context, collection string, data interface{}) error {
	c := m.db.Collection(collection)
	_, err := c.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}
func (m MongodbDao) PagingQuery(collection string, filter interface{}, limit int64, index int64, project bson.M) ([]map[string]interface{}, error) {
	var data []map[string]interface{}
	findoptions := &options.FindOptions{}
	if limit >= 0 {
		findoptions.SetLimit(limit)
		findoptions.SetSkip(limit * index)
	}
	if project != nil {
		findoptions.SetProjection(project)
	}
	c := m.db.Collection(collection)
	r, err := c.Find(context.TODO(), filter, findoptions)
	if err != nil {
		return nil, err
	}
	defer r.Close(context.TODO())
	for r.Next(context.TODO()) {
		var item map[string]interface{}
		err = r.Decode(&item)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	// r.Decode(&data)
	return data, nil
}

func (m MongodbDao) AggregateQuery(collection string, pipeline []bson.D) ([]bson.M, error) {
	// 执行聚合查询
	c := m.db.Collection(collection)
	cursor, err := c.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	// 打印出今天没有更新的appid
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (m MongodbDao) Find(ctx context.Context, collection string, filter interface{}, project interface{}) ([]map[string]interface{}, error) {
	var data []map[string]interface{}
	c := m.db.Collection(collection)
	r, err := c.Find(ctx, filter, options.Find().SetProjection(project))
	if err != nil {
		return nil, err
	}
	defer r.Close(ctx)
	for r.Next(ctx) {
		var item map[string]interface{}
		err = r.Decode(&item)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	return data, nil
}
