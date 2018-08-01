package main

import (
	"os"
	"fmt"
	_ "net/http/pprof"
	"net/http"

	"dormon.net/qq/web"

	"gopkg.in/urfave/cli.v1"
	"dormon.net/qq/mht"
)

var AppVersion = "0.0.1-dev"

func main() {

	// heat pprof
	go func() {
		http.ListenAndServe("0.0.0.0:10080", nil)
	}()

	app := cli.NewApp()
	app.Name = "QQGroupStatistics"
	app.Usage = "Just for fun!"
	app.Version = AppVersion
	app.Commands = []cli.Command{
		web.CMDRunWeb,
		mht.CMDRunImport,
		mht.CMDRunFix,
	}
	app.Flags = []cli.Flag{
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
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
