package es

import (
	"gopkg.in/urfave/cli.v1"
	"dormon.net/qq/utils"
	"dormon.net/qq/config"
	"github.com/sirupsen/logrus"
	"github.com/olivere/elastic"
	"context"
	"io/ioutil"
	"strings"
)

var CMDRunReindex = cli.Command{
	Name:        "reindex",
	Usage:       "reindex ES's index for new mapping",
	Description: "As modifying ES's mapping, reindex the index to apply new changes.",
	Flags: []cli.Flag{
		// 指定旧的index的名称，数据将从这个index导入到新的index，不能为空
		cli.StringFlag{
			EnvVar: "DORMON_QQ_OLD_INDEX_NAME",
			Name:   "old_index, i",
			Usage:  "specify the old index name, will dump data from this index to new one. This param can not be null",
			Value:  "",
		},
		// 指定新的index的名称，名称不能够与原有index重复
		cli.StringFlag{
			EnvVar: "DORMON_QQ_NEW_INDEX_NAME",
			Name:   "index, I",
			Usage:  "specify the new index name, default is a random name",
			Value:  "",
		},
		// 指定新的mapping的json文件路径，默认为./mapping.json
		cli.StringFlag{
			EnvVar: "DORMON_QQ_NEW_MAPPING",
			Name:   "mapping, M",
			Usage:  "specify the new mapping json file path",
			Value:  "./mapping.json",
		},
		// 是否删除旧的index, 默认不删除
		cli.BoolFlag{
			EnvVar: "DORMON_QQ_DELETE_OLD_INDEX",
			Name:   "delete_index, D",
			Usage:  "whether to delete old index",
		},
		// 是否在完成完成建立新index后切换到新的index，默认切换
		cli.BoolTFlag{
			EnvVar: "DORMON_QQ_SWITCH_TO_NEW_INDEX",
			Name:   "switch, S",
			Usage:  "whether to switch to new index after new index created.",
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
	Action: runReindex,
}

func runReindex(c *cli.Context) {
	config.InitialConfig(c)
	InitialES()

	esClient := ElasticClient()

	sourceIndex := c.String("old_index")
	if sourceIndex == "" {
		logrus.Fatalf("Old index name is required")
		return
	}

	targetIndex := c.String("index")
	if targetIndex == "" {
		targetIndex = strings.ToLower(utils.RandomString(16))
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

	createIndex(esClient, targetIndex, readMapping(mappingPath))

	src := elastic.NewReindexSource().Index(sourceIndex)
	dst := elastic.NewReindexDestination().Index(targetIndex)
	res, err := esClient.Reindex().Source(src).Destination(dst).Refresh("true").Do(context.Background())

	if err != nil {
		logrus.Fatalf("Reindex failed with error: %s", err)
	}

	logrus.Infof("Reindex finished, reindexed documents count %d in %d ms", res.Total, res.Took)

	if c.Bool("switch") {
		setIndexAlias(esClient, config.Config().ElasticSearchConfig.AliasName, targetIndex)
	}

	if c.Bool("delete_index") {
		deleteIndex(esClient, sourceIndex)
	}

}

func readMapping(path string) string {

	mapping, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatalf("Read mapping json file failed with error: %s", err)
	}

	return string(mapping)

}

func deleteIndex(client *elastic.Client, index string) {

	indicesDeleteResponse, err := client.DeleteIndex("twitter").Do(context.Background())

	if err != nil {
		logrus.Fatalf("Failed to delete index %s with error: %s", index, err)
	}

	if !indicesDeleteResponse.Acknowledged {
		logrus.Fatalf("Failed to delete index %s", index)
	}

	logrus.Infof("Index %s deleted", index)

}
