package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	app.Command("scan", "Perform scan of a given host", func(cmd *cli.Cmd) {
		addr := cmd.StringArg("IP_ADDR", "", "IP address of a host to scan")
		script := cmd.StringOpt("script", "", "Script to be executed")
		componentsRaw := cmd.StringOpt("components", "none", "Components to discover - possible values: none | all | eth,mem,fcc,cpu,disk")
		withBIOSAndFirmware := cmd.BoolOpt("with-bios-and-firmware", false, "Try to discover BIOS and firmware versions")
		withModel := cmd.BoolOpt("with-model", false, "Append detected model name to \"Remarks\" field in Ralph")
		dryRun := cmd.BoolOpt("dry-run", false, "Don't save anything in Ralph")

		cmd.Spec = "IP_ADDR --script=<script name> [--components=<comma-separated list of components>] [--with-bios-and-firmware] [--with-model] [--dry-run]"

		cmd.Action = func() {
			if *script == "" {
				log.Fatalln("No script supplied to '--script' switch. Aborting.")
			}
			// TODO(xor-xor): Consider adding some message when no --components
			// *and* --with-bios-and-firmware *and* --with-model is given, or
			// make at least one of them required.
			components, err := parseComponents(*componentsRaw)
			if err != nil {
				log.Fatalf("Error parsing value(s) for '--component' switch: %s. Aborting.", err)
			}
			if changesDetected := PerformScan(*addr, *script, *components, *withBIOSAndFirmware, *withModel, *dryRun, cfg, cfgDir); !changesDetected {
				log.Println("No changes detected.")
			}
		}
	})

	app.Version("v version", "0.1.0")
	app.Run(os.Args)
}

// parseComponents returns a map denoting presence or absence of a given
// component in --components=<...> switch. Aborts the program when an unknown
// component is found.
func parseComponents(componentsRaw string) (*map[string]bool, error) {
	var components = map[string]bool{
		"none": false,
		"all":  false,
		"eth":  false,
		"mem":  false,
		"fcc":  false,
		"cpu":  false,
		"disk": false,
	}
	cc := strings.Split(componentsRaw, ",")
	for _, c := range cc {
		if _, ok := components[c]; !ok {
			return nil, fmt.Errorf("unknown component: %s", c)
		}
		components[c] = true
	}
	if components["none"] == true && len(cc) > 1 {
		return nil, errors.New("invalid combination: \"none\" option should be used exclusively")
	}
	if components["all"] == true && len(cc) > 1 {
		return nil, errors.New("invalid combination: \"all\" option should be used exclusively")
	}
	return &components, nil
}
