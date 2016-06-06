package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jawher/mow.cli"
	gommonlog "github.com/labstack/gommon/log"
)

func main() {
	var w io.Writer
	log.SetFlags(0)

	cfgDir, err := GetCfgDirLocation("")
	if err != nil {
		fmt.Println(err)
	}
	err = PrepareCfgDir(cfgDir)
	if err != nil {
		fmt.Println(err)
	}

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

	app.Command("scan", "Perform scan of a given host/network", func(cmd *cli.Cmd) {
		addr := cmd.StringArg("ADDR", "", "Address of a host to scan (IP or FQDN)")
		scripts := cmd.StringsOpt("scripts", []string{"idrac.py"}, "Scripts to be executed")
		dryRun := cmd.BoolOpt("dry-run", false, "Don't write anything")

		cmd.Spec = "ADDR [--scripts=<scripts>] [--dry-run]"

		cmd.Action = func() {
			PerformScan(*addr, *scripts, *dryRun, cfgDir)
		}
	})

	app.Version("v version", "0.1.0")
	app.Run(os.Args)
}
