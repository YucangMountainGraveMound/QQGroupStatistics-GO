package es

import (
	"context"
	"sync"

	"dormon.net/qq/config"

	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
	"dormon.net/qq/errors"
)

var once sync.Once
var elasticClient *elastic.Client
var esIndex string

func ElasticClient() *elastic.Client {
	var err error
	once.Do(func() {
		var esUrl = "http://" + config.Config().ElasticSearchConfig.Host + ":" + config.Config().ElasticSearchConfig.Port
		esIndex = config.Config().ElasticSearchConfig.IndexName

		elasticClient, err = elastic.NewClient(
			elastic.SetURL(esUrl),
			elastic.SetSniff(false),
		)
		if err != nil {
			panic(err)
		}

		info, code, err := elasticClient.Ping(esUrl).Do(context.Background())
		if err != nil {
			panic(err)
		}
		logrus.Infof("Elasticsearch ping: code %d | version %s\n", code, info.Version.Number)

		esVersion, err := elasticClient.ElasticsearchVersion(esUrl)
		if err != nil {
			panic(err)
		}
		logrus.Infof("Elasticsearch version: %s\n", esVersion)
	})

	return elasticClient
}

func InitialES(c *cli.Context) error {

	esClient := ElasticClient()

	// 检查是否存在qq index
	exist, err := esClient.IndexExists(esIndex).Do(context.Background())
	if err != nil {
		panic(err)
	}

	mapping := `
{
	"settings": {
		"number_of_shards": 5,
		"number_of_replicas": 1
	},
	"mappings": {
		"qq": {
			"properties": {
				"number": {
					"type": "keyword"
				},
				"message": {
					"type": "text",
					"analyzer": "ik_max_word",
					"search_analyzer": "ik_max_word",
					"fielddata": true,
					"boost": 5,
					"term_vector": "with_positions_offsets"
				},
				"date": {
					"type": "date"
				},
				"images": {
					"type": "keyword"
				},
				"at": {
					"type": "keyword"
				},
				"interval": {
					"type": "integer"
				},
				"message_len": {
					"type": "integer"
				}
			}
		}
	}
}
`
	if !exist {
		createIndex, err := esClient.CreateIndex(esIndex).BodyString(mapping).Do(context.Background())
		if err != nil {
			panic(err)
		}
		if !createIndex.Acknowledged {
			logrus.Fatalf("create es index failed")
			return errors.New("create es index failed")
		}
	}
	return nil
}
