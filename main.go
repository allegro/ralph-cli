package main

import (
	"io"
	"log"
	"os"

	"github.com/jawher/mow.cli"
	gommonlog "github.com/labstack/gommon/log"
)

func main() {
	var w io.Writer
	log.SetFlags(0)
	cfg, _ := GetConfig()
	switch cfg.LogOutput {
	case "logstash":
		w = NewLogstashWriter(cfg)
	default:
		w = os.Stderr
	}
	log.SetOutput(w)
	gommonlog.SetOutput(w)

	app := cli.App("ralph-cli", "Command-line interface for Ralph")
	app.Spec = "IP"
	var ip = app.StringArg("IP", "", "IP address to scan")
	app.Action = func() {
		PerformDummyScan(ip)
	}
	app.Version("v version", "0.1.0")
	app.Run(os.Args)
}
