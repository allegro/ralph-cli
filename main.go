package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/jawher/mow.cli"
	gommonlog "github.com/labstack/gommon/log"
)

func main() {
	var w io.Writer
	log.SetFlags(0)

	cfgDir, err := GetCfgDirLocation("")
	if err != nil {
		log.Fatalln(err)
	}
	cfgFileName := "config.toml"
	err = PrepareCfgDir(cfgDir, cfgFileName)
	if err != nil {
		log.Fatalln(err)
	}
	cfg, err := GetConfig(filepath.Join(cfgDir, cfgFileName))
	if err != nil {
		log.Fatalln(err)
	}
	switch cfg.LogOutput {
	// TODO(xor-xor): Uncomment this when logstash implementation will be ready.
	// case "logstash":
	// 	w = NewLogstashWriter(cfg)
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
			PerformScan(*addr, *scripts, *dryRun, cfg, cfgDir)
		}
	})

	app.Version("v version", "0.1.0")
	app.Run(os.Args)
}
