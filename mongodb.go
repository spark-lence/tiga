package tiga

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongodbDao struct {
	db     *mongo.Database
	config Configuration
}
type MongoStruct interface{
	Save()error
	Read()interface{}
}
func NewMongodbDao(config Configuration) *MongodbDao {
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
		db: client.Database(db),
		config: config,
	}

}

