package mht

import (
	"bufio"
	"io/ioutil"
	"os"

	"dormon.net/qq/config"
	"dormon.net/qq/db"
	"dormon.net/qq/es"
	"dormon.net/qq/web/model"

	"gopkg.in/urfave/cli.v1"
)

var CMDRunImport = cli.Command{
	Name:   "import",
	Usage:  "import the record",
	Action: RunImport,
	Flags: []cli.Flag{
		cli.StringFlag{
			EnvVar: "DORMON_QQ_RECORD",
			Name:   "record, F",
			Usage:  "specify the configuration file",
			Value:  "./record.mht",
		},
		cli.StringFlag{
			EnvVar: "DORMON_QQ_CONFIG",
			Name:   "config, C",
			Usage:  "specify the configuration file",
			Value:  "./config.toml",
		},
		cli.BoolFlag{
			EnvVar: "DORMON_QQ_OVERWRITE",
			Name:   "overwrite, O",
			Usage:  "overwrite the record",
		},
	},
}

var CMDRunFix = cli.Command{
	Name:   "fix",
	Usage:  "standardizing the record file",
	Action: RunFix,
	Flags: []cli.Flag{
		cli.StringFlag{
			EnvVar: "DORMON_QQ_RECORD",
			Name:   "record, F",
			Usage:  "specify the configuration file",
			Value:  "./record.mht",
		},
	},
}

func RunImport(c *cli.Context) {

	config.InitialConfig(c)

	es.InitialES()

	db.InitialDB()
	autoMigrate()

	if c.Bool("overwrite") {
		db.RedisClient().Do("flushall")
	}

	content, err := ioutil.ReadFile(c.String("record"))

	mht := New()

	if err != nil {
		panic(err)
	}

	err = mht.Parse(content)

	if err != nil {
		panic(err)
	}

}

func RunFix(c *cli.Context) {
	// fixRecordMhtFile 文档格式有些问题，导致无法正确分割mht文件，修正
	f, err := os.OpenFile(c.String("record"), os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Seek(132, 0)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 1)
	_, err = f.Read(b)
	if err != nil {
		panic(err)
	}
	if string(b) != ";" {
		_, err = f.Seek(132, 0)
		if err != nil {
			panic(err)
		}
		w := bufio.NewWriter(f)
		_, err = w.WriteString(";\n")
		if err != nil {
			panic(err)
		}
		w.Flush()
	}

}

// 自动创建表结构
func autoMigrate() {
	db.GetDB().AutoMigrate(
		model.User{},
		model.Image{},
		model.Dictionary{},
	)
}
