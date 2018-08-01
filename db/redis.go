package db

import (
	"sync"

	"dormon.net/qq/config"

	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
)

var once sync.Once
var redisClient redis.Conn

func RedisClient() redis.Conn {
	var err error
	once.Do(func() {
		redisClient, err = redis.Dial("tcp", config.Config().RedisConfig.Host+":"+config.Config().RedisConfig.Port)
		if err != nil {
			logrus.Fatalf("Error when connecting Redis server with Error: %s", err)
			return
		}
	})

	return redisClient
}
