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
	"encoding/json"
	"github.com/jinzhu/gorm"
)

type Record struct {
	Number      string    `json:"number"`
	Message     string    `json:"message"`
	Date        time.Time `json:"date"`
	Images      []string  `json:"images"`
	At          string    `json:"at"`
	Interval    int64     `json:"interval"`
	MessageLen  int       `json:"message_len"`
	Expression  string    `json:"expression"`
	MessageType string    `json:"message_type"`
}

type Expression struct {
	gorm.Model
	Key   string `gorm:"type:varchar(10);unique_index"`
	Value string `gorm:"type:varchar(200)"`
}

type Message struct {
	ClassName string `json:"className"`
	FriendUin string `json:"friendUin"`
	Message   string `json:"message"`
	SelfUin   string `json:"selfUin"`
	SenderUin string `json:"senderUin"`
	Time      string `json:"time"`
	UniSeq    string `json:"uniSeq"`
}

type Picture struct {
	PicUrl string `json:"picUrl"`
	UniSeq string `json:"uniSeq"`
}

func CreateRecordFromXposedMessage(message Message) error {

	index := config.Config().ElasticSearchConfig.AliasName
	esClient := es.ElasticClient()

	dateTime, err := dateparse.ParseAny(message.Time)
	if err != nil {
		logrus.Errorf("Failed to parse dateTime %s", message.Time)
		return err
	}

	record := &Record{
		Number:      message.SenderUin,
		Date:        dateTime,
		Message:     message.Message,
		MessageLen:  strings.Count(message.Message, ""),
		MessageType: message.ClassName,
	}

	d := getDocument(message.UniSeq)
	if d == nil {
		// 如果没有记录就直接创建
		record.trim()
		_, err := esClient.Index().Index(index).Id(message.UniSeq).Type(index).BodyJson(record).Do(context.Background())
		if err != nil {
			logrus.Errorf("Failed to create document id %s with error: %s", message.UniSeq, err)
			return err
		}
		logrus.Debugf("Message doc id %s created", message.UniSeq)
	} else {
		// 如果存在就获取内容
		source, err := d.Source.MarshalJSON()
		if err != nil {
			logrus.Errorf("Failed to marshal source to json with error: %s", err)
		}
		var r Record
		err = json.Unmarshal(source, &r)
		if err != nil {
			logrus.Errorf("Failed to marshal source json with error: %s", err)
		}
		// 通过number字段判断message是否已经创建，若为空，表示该记录是由picture创建的
		if r.Number == "" {
			record.trim()
			_, err := esClient.Update().Index(config.Config().ElasticSearchConfig.AliasName).Type(config.Config().ElasticSearchConfig.AliasName).Id(message.UniSeq).Doc(map[string]interface{}{
				"number":       record.Number,
				"date":         record.Date,
				"message":      record.Message,
				"message_len":  record.MessageLen,
				"message_type": record.MessageType,
			}).Do(context.Background())
			if err != nil {
				logrus.Errorf("Failed to update msg, document id %s with error: %s", message.UniSeq, err)
				return err
			}
			logrus.Debugf("Message doc id %s updated", message.UniSeq)
		} else {
			logrus.Debugf("Document %s already exists, skip.", message.UniSeq)
		}
	}

	return nil
}

func CreateRecordFromXposedPicture(picture Picture) error {
	index := config.Config().ElasticSearchConfig.AliasName
	esClient := es.ElasticClient()

	d := getDocument(picture.UniSeq)

	hash, err := utils.DownloadImageToBase64(picture.PicUrl)
	if err != nil {
		logrus.Errorf("Failed to download pic %s with error: %s", picture.PicUrl, err)
	}

	if d == nil {
		// 没有记录的情况
		record := &Record{
			Images: []string{hash},
		}
		_, err := esClient.Index().Index(index).Id(picture.UniSeq).Type(index).BodyJson(record).Do(context.Background())
		if err != nil {
			logrus.Errorf("Failed to create document id %s with error: %s", picture.UniSeq, err)
			return err
		}
		logrus.Infof("Picture doc id %s created", picture.UniSeq)
	} else {
		// 有记录的情况先获取doc
		source, err := d.Source.MarshalJSON()
		if err != nil {
			logrus.Errorf("Failed to marshal source to json with error: %s", err)
		}
		var r Record
		err = json.Unmarshal(source, &r)
		if err != nil {
			logrus.Errorf("Failed to marshal source json with error: %s", err)
		}
		if len(r.Images) != 0 {
			// 检查图片是否已经存在
			for i := 0; i < len(r.Images); i ++ {
				if r.Images[i] == hash {
					// 图片已经存在，退出
					logrus.Debugf("Picture hash %s exists", hash)
					return nil
				}
			}
		}

		// 不存在则更新
		imageArr := append(r.Images, hash)
		_, err = esClient.Update().Index(config.Config().ElasticSearchConfig.AliasName).Type(config.Config().ElasticSearchConfig.AliasName).Id(picture.UniSeq).Doc(map[string]interface{}{
			"images": imageArr,
			"date":   time.Now(),
		}).Do(context.Background())
		if err != nil {
			logrus.Errorf("Failed to update pic, document id %s with error: %s", picture.UniSeq, err)
			return err
		}
		logrus.Debugf("Picture doc id %s updated", picture.UniSeq)
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

	// 消息存在不正确的时间
	if record.Date.After(time.Now()) {
		record.Number = ""
	}

	// 计算消息间隔
	// TODO:改为定时任务遍历所有文档来计算

	// 处理@谁
	if record.At != "" {
		record.At = getAccount(record.At)
	}

	// TODO:目前没有好的处理@xxx之后没有空格的情况
	if strings.Contains(record.Message, "@") {
		a := getAccount(strings.Split(strings.Split(record.Message, "@")[1], " ")[0])
		if a != "" {
			record.At = a
		}
	}

	// 表情处理
	if len(record.Message) < 2 {
		bIndex := strings.Index(record.Message, "\x14")
		if bIndex != -1 {
			// 存在表情
			// 表情占长度可能占2-3个字节
			if len(record.Message[bIndex:]) >= 2 {
				exp := record.Message[bIndex : bIndex+2]
				if config.GetExpression()[exp] != "" {
					record.Expression = config.GetExpression()[exp] + ".png"
					record.Message = strings.Replace(record.Message, exp, "", -1)
				} else {
					if len(record.Message[bIndex:]) >= 3 {
						exp = record.Message[bIndex : bIndex+3]
						if config.GetExpression()[exp] != "" {
							record.Expression = config.GetExpression()[exp] + ".png"
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
