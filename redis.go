package tiga


import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDao struct {
	client *redis.Client
	config Configuration
}

// NewRdbConfig redis 配置构造函数
func NewRdbConfig(config Configuration) *redis.Options {
	env := config.GetEnv()
	password := config.GetConfigByEnv(env, "redis.password").(string)
	username := config.GetConfigByEnv(env, "redis.username").(string)
	addr := config.GetString(fmt.Sprintf("%s.%s", env, "redis.addr"))
	db := config.GetInt(fmt.Sprintf("%s.%s", env, "redis.db"))

	connectSize := config.GetInt(fmt.Sprintf("%s.%s", env, "redis.connectSize"))
	timeout := config.GetInt(fmt.Sprintf("%s.%s", env, "redis.timeout")) * int(time.Second)
	return &redis.Options{
		Addr:     addr,
		Password: password, // 密码
		Username: username, //用户名
		DB:       db,
		//连接池容量及闲置连接数量
		PoolSize:     connectSize, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: 10,          //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

		//超时
		DialTimeout:  time.Duration(timeout), //连接建立超时时间，默认5秒。
		ReadTimeout:  time.Duration(timeout), //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: time.Duration(timeout), //写超时，默认等于读超时
		PoolTimeout:  time.Duration(timeout), //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		//命令执行失败时的重试策略
		MaxRetries:      3,                      // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   //每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, //每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

	}
}

func NewRedisDao(config Configuration) *RedisDao {
	client := redis.NewClient(NewRdbConfig(config))
	err := client.Ping(context.TODO()).Err()
	if err != nil {
		panic(err)
	}
	return &RedisDao{
		client: client,
	}
}

func (r *RedisDao) BFAdd(key string, value string) bool {
	inserted, err := r.client.Do(context.Background(), "BF.ADD", key, value).Bool()
	if err != nil {
		panic(err)
	}
	if inserted {
		return true
	} else {
		return false
	}
}
func (r *RedisDao) Set(key string, value interface{}, expiration time.Duration) error {
	key = fmt.Sprintf("%s:%s", key, r.config.GetEnv())
	err := r.client.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis set %s to %v error %w", key, value, err)
	}
	return nil
}
func (r *RedisDao) Get(key string) string {
	key = fmt.Sprintf("%s:%s", key, r.config.GetEnv())
	val := r.client.Get(context.Background(), key).Val()

	return val
}
func (r *RedisDao) GetInt(key string) (int, error) {
	key = fmt.Sprintf("%s:%s", key, r.config.GetEnv())
	val, err := r.client.Get(context.Background(), key).Int()

	return val, err
}

func (r *RedisDao) SetNX(key string, val interface{}, expiration time.Duration) (bool, error) {
	key = fmt.Sprintf("%s:%s", key, r.config.GetEnv())
	ok, err := r.client.SetNX(context.Background(), key, val, expiration).Result()

	return ok, err
}
