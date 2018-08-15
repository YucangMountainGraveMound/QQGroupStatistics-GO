package model

import (
	"context"
	"strings"
	"time"

	"dormon.net/qq/config"
	"dormon.net/qq/es"
	"dormon.net/qq/utils"

	"github.com/sirupsen/logrus"
	"github.com/olivere/elastic"
	"github.com/araddon/dateparse"
	"strconv"
	"encoding/json"
	"dormon.net/qq/db"
	"github.com/garyburd/redigo/redis"
)

type Record struct {
	Number      string    `json:"number"`
	Message     string    `json:"message"`
	Date        time.Time `json:"date"`
	Images      []string  `json:"images"`
	At          string    `json:"at"`
	Interval    int64     `json:"interval"`
	MessageLen  int       `json:"message_len"`
	Expression  []string  `json:"expression"`
	MessageType string    `json:"message_type"`
}

type CoolQRecord struct {
	Anonymous   CoolQAnonymous `json:"anonymous"`
	Font        int            `json:"font"`
	GroupId     int            `json:"group_id"`
	Message     []CoolQMessage `json:"message"`
	MessageId   int            `json:"message_id"`
	MessageType string         `json:"message_type"`
	PostType    string         `json:"post_type"`
	RawMessage  string         `json:"raw_message"`
	SelfId      int            `json:"self_id"`
	SubType     string         `json:"sub_type"`
	Time        int            `json:"time"`
	UserId      int            `json:"user_id"`
}

type CoolQMessage struct {
	Data        CoolQMessageItem `json:"data"`
	MessageType string           `json:"type"`
}

type CoolQMessageItem struct {
	Text string `json:"text"`
	File string `json:"file"`
	Url  string `json:"url"`
	Id   string `json:"id"`
	QQ   string `json:"qq"`
}

type CoolQAnonymous struct {
	Flag string `json:"flag"`
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func CreateRecordFromCoolQMessage(CQrecord CoolQRecord) error {

	index := config.Config().ElasticSearchConfig.AliasName
	esClient := es.ElasticClient()

	dateTime, err := dateparse.ParseIn(strconv.Itoa(CQrecord.Time), time.UTC)
	dateTime = dateTime.Add(8 * time.Hour)
	if err != nil {
		logrus.Errorf("Failed to parse dateTime %s", CQrecord.Time)
		return err
	}

	record := &Record{
		Number: strconv.Itoa(CQrecord.UserId),
		Date:   dateTime,
	}

	for _, v := range CQrecord.Message {
		if v.MessageType == "text" {
			record.Message += v.Data.Text
		}
		if v.MessageType == "image" {
			record.Images = append(record.Images, v.Data.File)
			go utils.DownloadImage(v.Data.File, v.Data.Url)
		}
		if v.MessageType == "record" {
			//TODO:音频文件处理
		}
		if v.MessageType == "bface" {
			//TODO:原创表情处理
		}
		if v.MessageType == "at" {
			record.At = v.Data.QQ
		}
		if v.MessageType == "face" {
			record.Expression = append(record.Expression, v.Data.Id+".png")
		}
	}

	record.MessageLen = strings.Count(record.Message, "")

	lastRecordTime, err := redis.Int64(db.RedisConn().Do("GET", "lastRecord_"+record.Number))
	if err == redis.ErrNil {
		lastRecordTime = 0
	} else if err != nil {
		logrus.Errorf("Error get lastRecordTime from redis with error: %s", err)
	}

	if lastRecordTime == 0 {
		if getLastRecordTime(record.Number) == 0 {
			lastRecordTime = record.Date.Unix()
			record.Interval = -1
		} else {
			lastRecordTime = getLastRecordTime(record.Number)
			record.Interval = record.Date.Unix() - lastRecordTime
		}
	} else {
		lastRecordTime = getLastRecordTime(record.Number)
		record.Interval = record.Date.Unix() - lastRecordTime
	}

	_, err = db.RedisConn().Do("SET", "lastRecord_"+record.Number, record.Date.Unix())
	if err != nil {
		logrus.Errorf("Error set lastRecordTime from redis with error: %s", err)
	}

	_, err = esClient.Index().Index(index).Type(index).BodyJson(record).Do(context.Background())
	if err != nil {
		logrus.Errorf("Failed to create document id %s with error: %s", record, err)
		return err
	}

	return nil
}

func CreateRecordFromImport(record *Record) {

	if record.trim(); record.Number == "" {
		// config中若不存在对应的qq号码或者昵称，则跳过
		return
	}

	esClient := es.ElasticClient()

	index := config.Config().ElasticSearchConfig.AliasName
	recordHash := recordHash(record)

	d := getDocument(recordHash)

	if d == nil {
		_, err := esClient.Index().Index(index).Id(recordHash).Type(index).BodyJson(record).Do(context.Background())
		if err != nil {
			logrus.Errorf("Failed to create document id %s with error: %s", recordHash, err)
		}
		logrus.Debugf("Record created. Record id: %s", recordHash)
	} else {
		logrus.Debugf("Record already exist, skip! Record id: %s", recordHash)
	}
}

// 消息内容处理
func (record *Record) trim() {
	// 账号处理
	record.Number = getAccount(record.Number)

	// 消息内容处理
	// 空消息处理
	if record.MessageLen == 0 {
		record.Message = ""
	}

	// 图片消息处理
	record.Message = strings.Replace(record.Message, "[图片]", "", -1)

	// 空格处理
	record.Message = strings.Replace(record.Message, "\u00a0", " ", -1)

	// 计算消息长度
	record.MessageLen = strings.Count(record.Message, "")

	// 系统消息处理
	if record.Number == "10000" {
		record.Message = trimSystemMessage(record.Message, "422680319")
	}

	// 计算消息间隔
	// 改为定时任务遍历所有文档来计算

	// 处理@谁
	if record.At != "" {
		record.At = getAccount(record.At)
	}

	// TODO:目前没找到好的处理@xxx之后没有空格的情况的方法
	if strings.Contains(record.Message, "@") {
		a := getAccount(strings.Split(strings.Split(record.Message, "@")[1], " ")[0])
		if a != "" {
			record.At = a
		}
	}

	// 表情处理
	if len(record.Message) >= 2 {
		bIndex := strings.Index(record.Message, "\x14")
		if bIndex != -1 {
			// 存在表情
			// 表情占长度可能占2-3个字节
			if len(record.Message[bIndex:]) >= 2 {
				exp := record.Message[bIndex : bIndex+2]
				if config.GetExpression()[exp] != "" {
					record.Expression = append(record.Expression, config.GetExpression()[exp]+".png")
					record.Message = strings.Replace(record.Message, exp, "", -1)
				} else {
					if len(record.Message[bIndex:]) >= 3 {
						exp = record.Message[bIndex : bIndex+3]
						if config.GetExpression()[exp] != "" {
							record.Expression = append(record.Expression, config.GetExpression()[exp]+".png")
							record.Message = strings.Replace(record.Message, exp, "", -1)
						}
					}
				}
			}
		}
	}
}

// trimSystemMessage 处理系统消息
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

// recordHash 计算消息的hash
func recordHash(record *Record) string {
	return utils.MD5(record.Number + record.Date.String() + record.Message)
}

// getAccount 通过昵称查找发送消息的账号
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
		return ""
	}
	return account
}

// getDocument 检查某个消息是否已经存在
func getDocument(id string) *elastic.GetResult {
	esClient := es.ElasticClient()
	res, err := esClient.Get().Index(config.Config().ElasticSearchConfig.AliasName).Type(config.Config().ElasticSearchConfig.AliasName).Id(id).Do(context.TODO())
	if err != nil {
		return nil
	}
	return res
}

func getDocumentField(id, field string) *elastic.GetResult {
	esClient := es.ElasticClient()
	fsc := elastic.NewFetchSourceContext(true).Exclude(field)
	res, err := esClient.Get().Index(config.Config().ElasticSearchConfig.AliasName).Type(config.Config().ElasticSearchConfig.AliasName).Id(id).FetchSourceContext(fsc).Do(context.TODO())
	if err != nil {
		return nil
	}
	return res
}

/**
{
  "size": 1,
  "query": {
    "term": {
      "number": "422680319"
    }
  },
  "sort": {
    "date": {
      "order": "desc"
    }
  }
}
 */
func getLastRecordTime(number string) int64 {
	esClient := es.ElasticClient()
	query := elastic.NewTermQuery("number", number)
	result, err := esClient.Search().Index(config.Config().ElasticSearchConfig.AliasName).Query(query).Sort("date", false).Size(1).Do(context.Background())
	if err != nil {
		return 0
	}

	if len(result.Hits.Hits) == 0 {
		return 0
	}

	var record Record
	err = json.Unmarshal(*result.Hits.Hits[0].Source, &record)
	if err != nil {
		return 0
	}

	return record.Date.Unix()
}
