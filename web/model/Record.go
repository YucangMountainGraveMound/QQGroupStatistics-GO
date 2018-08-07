package model

import (
	"context"
	"strings"
	"time"

	"dormon.net/qq/config"
	"dormon.net/qq/db"
	"dormon.net/qq/es"
	"dormon.net/qq/utils"

	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
)

type Record struct {
	Number     string    `json:"number"`
	Message    string    `json:"message"`
	Date       time.Time `json:"date"`
	Images     []string  `json:"images"`
	At         string    `json:"at"`
	Interval   int64     `json:"interval"`
	MessageLen int       `json:"message_len"`
}

func CreateRecord(record *Record) {

	if record.trim(); record.Number == "" {
		// config中若不存在对应的qq号码或者昵称，则跳过
		return
	}

	esClient := es.ElasticClient()
	redisClient := db.RedisClient()

	index := config.Config().ElasticSearchConfig.IndexName

	dataExist, err := redis.Bool(redisClient.Do("EXISTS", recordHash(record)))

	if err != nil {
		logrus.Errorf("Error when creating record %s with error: %s", record, err)
	}

	if !dataExist {
		_, err := esClient.Index().Index(index).Type(index).BodyJson(record).Do(context.Background())
		if err != nil {
			logrus.Errorf("Error when Indexing %s with error: %s", index, err)
		}

		_, err = redisClient.Do("SET", recordHash(record), "")
		if err != nil {
			logrus.Errorf("Redis Error when doing SET record hash: %s", err)
		}

		_, err = redisClient.Do("SET", record.Number, record.Date.Unix())
		if err != nil {
			logrus.Errorf("Redis Error when doing SET record date: %s", err)
		}

		if err != nil {
			panic(err)
		}
		logrus.Infof("Record created. Record time: %s", record.Date)
	} else {
		logrus.Infof("Record already exist, skip! Record time: %s", record.Date)
	}
}

func (record *Record) trim() {
	// 账号处理
	record.Number = getAccount(record.Number)

	// 消息内容处理
	// 空消息处理
	if record.MessageLen == 0 {
		record.Message = ""
	}
	// 空格处理
	record.Message = strings.Replace(record.Message, "\u00a0", " ", -1)

	// 系统消息处理
	if record.Number == "10000" {
		record.Message = trimSystemMessage(record.Message, "422680319")
	}

	// 消息存在不正确的时间
	if record.Date.After(time.Now()) {
		record.Number = ""
	}

	// 计算消息间隔
	redisClient := db.RedisClient()
	lastRecordUnixTime, err := redis.Int64(redisClient.Do("GET", record.Number))
	if err != nil {
		logrus.Errorf("Redis Error when doing GET record date: %s", err)
		lastRecordUnixTime = 0
	}
	record.Interval = record.Date.Unix() - lastRecordUnixTime

	// 处理@谁
	if record.At != "" {
		record.At = getAccount(record.At)
	}
}

func trimSystemMessage(message, operator string) string {

	// 这里需要判断你、您的对象是谁，即是谁上传的

	// 撤回消息
	if str := "撤回了一条消息"; strings.Contains(message, str) {
		user := strings.Split(message, str)[0]
		if user == "你" {
			return operator + "|" + message
		} else {
			return getAccount(user) + "|" + message
		}
	}

	// 修改群聊主题
	if str := "修改群聊的主题为"; strings.Contains(message, str) {
		user := strings.Split(message, str)[0]
		if user == "您" {
			return operator + "|" + message
		} else {
			return getAccount(user) + "|" + message
		}
	}

	return ""
}

func recordHash(record *Record) string {
	return utils.MD5(record.Number + record.Date.String() + record.Message)
}

func getAccount(a string) string {
	found := false
	account := ""
	for _, v := range config.Config().Account {
		v.Account = strings.Replace(v.Account, "【管理员】", "", -1)
		if v.Account == a {
			account = v.Account
			found = true
			break
		} else {
			for _, v1 := range v.Alias {
				if v1 == a {
					found = true
				}
			}
			if found {
				account = v.Account
				break
			}
		}
	}
	if !found {
		logrus.Warnf("Cannot locate qq number in configs: [%s]", a)
	}
	return account
}
