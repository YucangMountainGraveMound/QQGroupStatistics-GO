package es

import (
	"context"
	"sync"

	"dormon.net/qq/config"

	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"time"
)

var once sync.Once
var elasticClient *elastic.Client
var esIndex string
var p *elastic.BulkProcessor

func ElasticClient() *elastic.Client {
	var err error
	once.Do(func() {
		var esUrl = "http://" + config.Config().ElasticSearchConfig.Host + ":" + config.Config().ElasticSearchConfig.Port
		esIndex = config.Config().ElasticSearchConfig.AliasName

		elasticClient, err = elastic.NewClient(
			elastic.SetURL(esUrl),
			elastic.SetSniff(false),
		)
		if err != nil {
			panic(err)
		}

		info, code, err := elasticClient.Ping(esUrl).Do(context.Background())
		if err != nil {
			logrus.Fatalf("Failed to connect to elasticsearch cluster with error: %s", err)
		}
		logrus.Infof("Elasticsearch ping: code %d | version %s", code, info.Version.Number)

		esVersion, err := elasticClient.ElasticsearchVersion(esUrl)
		if err != nil {
			logrus.Fatalf("Failed to get elasticsearch version with error: %s", err)
		}
		logrus.Infof("Elasticsearch version: %s", esVersion)
	})

	return elasticClient
}

func InitialES() error {

	esClient := ElasticClient()

	// 检查是否存在qq index
	exist, err := esClient.IndexExists(esIndex).Do(context.Background())
	if err != nil {
		logrus.Fatalf("Failed to check elasticsearch with error: %s", err)
	}

	if !exist {
		logrus.Fatalf("No found index [%s], do you create a index?")
	}

	return nil
}

func GetBulkProcess() (*elastic.BulkProcessor) {
	var err error
	esClient := ElasticClient()
	if p == nil {
		p, err = esClient.
			BulkProcessor().
			Name("BULK_PROCESS").
			Workers(4).
			BulkActions(1000).
			FlushInterval(10 * time.Second).
			Before(func(executionId int64, requests []elastic.BulkableRequest) {
				for v := range requests {
					logrus.Info(v)
				}
			}).
			After(func(executionId int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
				logrus.Errorf("Some error happened in the bulk process: %s", err.Error())
			}).Do(context.Background())
		if err != nil {
			logrus.Fatalf("Failed to create bulk process with error: %s", err)
			return nil
		}
	}
	return p
}
