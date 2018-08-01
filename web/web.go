package web

import (
	"dormon.net/qq/config"
	"dormon.net/qq/db"
	"dormon.net/qq/es"
	"dormon.net/qq/web/model"

	"gopkg.in/urfave/cli.v1"
)

var CMDRunWeb = cli.Command{
	Name:        "web",
	Usage:       "run web",
	Description: "The website of qq statistics",
	Action:      runWeb,
}

// RunWeb run to serve a web application
func runWeb(c *cli.Context) {

	config.InitialConfig(c)

	es.InitialES(c)

	db.InitialDB()
	autoMigrate()

	model.InitialAccount()

	//InitialRouter().RunTLS("0.0.0.0:443", "./tls/214873497980883.pem", "./tls/214873497980883.key")
	InitialRouter().Run(":4000")
}

// 自动创建表结构
func autoMigrate() {
	db.GetDB().AutoMigrate(
		model.User{},
		model.Image{},
		model.Dictionary{},
	)
}
