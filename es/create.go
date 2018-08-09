package es

import (
	"gopkg.in/urfave/cli.v1"
	"github.com/sirupsen/logrus"
	"github.com/olivere/elastic"
	"context"
	"dormon.net/qq/config"
	"dormon.net/qq/utils"
	"strings"
)

var CMDRunCreate = cli.Command{
	Name:        "create",
	Usage:       "create a ES index",
	Description: "App require a index named qq or a index alias qq to run",
	Flags: []cli.Flag{
		// 指定要创建新的index的名称
		cli.StringFlag{
			EnvVar: "DORMON_QQ_NEW_INDEX_NAME",
			Name:   "index, I",
			Usage:  "specify the new index name to switch",
			Value:  "",
		},
		// 指定新的mapping的json文件路径，默认为./mapping.json
		cli.StringFlag{
			EnvVar: "DORMON_QQ_NEW_MAPPING",
			Name:   "mapping, M",
			Usage:  "specify the new mapping json file path",
			Value:  "./mapping.json",
		},
		cli.StringFlag{
			EnvVar: "DORMON_QQ_CONFIG",
			Name:   "config, C",
			Usage:  "specify the configuration file",
			Value:  "./config.toml",
		},
		cli.BoolFlag{
			EnvVar: "DORMON_QQ_GENERATE_CONFIG",
			Name:   "generate, G",
			Usage:  "generate a configuration file",
		},
	},
	Action: RunCreate,
}

func RunCreate(c *cli.Context) {
	config.InitialConfig(c)

	indexName := c.String("index")

	if indexName == "" {
		indexName = strings.ToLower(utils.RandomString(16))
	}

	mappingPath := c.String("mapping")
	if mappingPath == "" {
		logrus.Fatalf("Mapping json file is required")
	}

	exists, err := utils.PathExists(mappingPath, false)
	if err != nil {
		logrus.Fatalf("Checking file %s error: %s", mappingPath, err)
		return
	}
	if !exists {
		logrus.Fatalf("Not found mapping json file")
		return
	}

	createIndex(ElasticClient(), indexName, readMapping(mappingPath))
}

func createIndex(client *elastic.Client, index, mapping string) {

	indicesCreateResult, err := client.CreateIndex(index).BodyString(mapping).Do(context.Background())

	if err != nil {
		logrus.Fatalf("Failed to create index %s with error: %s", index, err)
	}

	if !indicesCreateResult.Acknowledged {
		logrus.Fatalf("Failed to create index %s", index)
	}

	logrus.Infof("Index %s created", index)

}
