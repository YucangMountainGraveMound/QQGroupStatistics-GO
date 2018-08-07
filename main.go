package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"dormon.net/qq/mht"
	"dormon.net/qq/web"

	"gopkg.in/urfave/cli.v1"
	"dormon.net/qq/es"
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
		es.CMDRunReindex,
		es.CMDRunCreate,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
