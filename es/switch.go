package es

import (
	"gopkg.in/urfave/cli.v1"
	"github.com/olivere/elastic"
	"context"
	"github.com/sirupsen/logrus"
	"dormon.net/qq/config"
)

var CMDRunSwitchIndex = cli.Command{
	Name:        "switch",
	Usage:       "set alias to a index",
	Description: "As modifying ES's mapping, reindex the index to apply new changes.",
	Flags: []cli.Flag{
		// 指定要切换的新的index的名称
		cli.StringFlag{
			EnvVar: "DORMON_QQ_NEW_INDEX_NAME",
			Name:   "index, I",
			Usage:  "specify the new index name to switch",
			Value:  "",
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
	Action: runSwitch,
}

func runSwitch(c *cli.Context) {
	config.InitialConfig(c)

	indexName := c.String("index")

	setIndexAlias(ElasticClient(), config.Config().ElasticSearchConfig.AliasName, indexName)
}

func setIndexAlias(client *elastic.Client, alias, index string) {

	aliasCreate, err := client.Alias().
		Add(index, alias).
		Do(context.TODO())

	if err != nil {
		logrus.Fatalf("Creating alias failed with error: %s", err)
	}

	if !aliasCreate.Acknowledged {
		logrus.Fatalf("Not creating new alias")
	}

	logrus.Infof("Success switch to %s index", index)

}
