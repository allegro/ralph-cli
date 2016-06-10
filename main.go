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
		addr := cmd.StringArg("IP_ADDR", "", "IP address of a host to scan")
		script := cmd.StringOpt("script", "", "Script to be executed")
		dryRun := cmd.BoolOpt("dry-run", false, "Don't write anything")

		cmd.Spec = "IP_ADDR --script=<script_name> [--dry-run]"

		cmd.Action = func() {
			if *script == "" {
				log.Fatalln("No script supplied to '--script' switch. Aborting.")
			}
			PerformScan(*addr, *script, *dryRun, cfg, cfgDir)
		}
	})

	app.Version("v version", "0.1.0")
	app.Run(os.Args)
}
