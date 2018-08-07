package db

import (
	"sync"

	"dormon.net/qq/config"

	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin/json"
	"bytes"
	"dormon.net/qq/errors"
)

var once sync.Once
var redisPool *RedisPool

type RedisPool struct {
	pool *redis.Pool
}

func newRedisPool() (*RedisPool, error) {

	pool := &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: time.Minute,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Config().RedisConfig.Host+":"+config.Config().RedisConfig.Port)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &RedisPool{pool}, nil
}

func RedisClient() *RedisPool {
	var err error

	once.Do(func() {
		redisPool, err = newRedisPool()
		if err != nil {
			logrus.Fatalf("Failed to connect to redis server with error: %s", err)
		}
	})

	return redisPool
}

// Enqueue 入队
func (client *RedisPool) Enqueue(queueName string, item interface{}) error {

	c := client.pool.Get()
	defer c.Close()

	jsonString, err := json.Marshal(item)
	if err != nil {
		return err
	}

	_, err = c.Do("rpush", queueName, jsonString)
	if err != nil {
		return err
	}

	return nil
}

// Pop 弹出一个元素
func (client *RedisPool) Pop(queueName string) (interface{}, error) {

	count, err := client.QueuedItemCount(queueName)
	if err != nil || count == 0 {
		return nil, errors.New("No item in queue [" + queueName + "]")
	}

	c := client.pool.Get()
	defer c.Close()

	results := make([]interface{}, int(count))
	var result interface{}

	for i := 0; i < int(count); i++ {
		reply, err := c.Do("LPOP", queueName)
		if err != nil {
			return nil, err
		}

		decoder := json.NewDecoder(bytes.NewReader(reply.([]byte)))
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}

		results = append(results, result)
	}
	return result, nil
}

// QueuedItemCount 队列长度
func (client *RedisPool) QueuedItemCount(queueName string) (int, error) {

	length, err := client.pool.Get().Do("llen", queueName)
	if err != nil {
		return 0, err
	}

	count, _ := length.(int64)

	return int(count), nil
}
