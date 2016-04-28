package cmd

import (
	"github.com/jawher/mow.cli"
	"github.com/quamilek/ralph-scan/scan"
)

type CLI struct {
	*cli.Cli
}

func ScanCli() *CLI {

	scanCli := &CLI{cli.App("ralph-scan", "Easy way to scan your devices")}

	scanCli.Spec = "[IP... [--components] [--plugins] [--dry-run]]"

	var (
		ip         = scanCli.StringsArg("IP", nil, "IP or HOSTNAME device to scan")
		components = scanCli.Strings(cli.StringsOpt{
			Name:  "components",
			Value: nil,
			Desc: `List components to scan. Available components:
                    CPU, RAM, DISK, ETHERNETS, DISK-SHARES`,
			EnvVar: "RALPH_SCAN_COMPONENTS",
		})
		plugins = scanCli.Strings(cli.StringsOpt{
			Name:   "plugins",
			Value:  nil,
			Desc:   "Ralph scan plugins to run. Available plugins: PLUGIN1, PLUGIN2",
			EnvVar: "RALPH_SCAN_PLUGINS",
		})
		dryRun = scanCli.Bool(cli.BoolOpt{
			Name:   "dry-run",
			Value:  false,
			Desc:   "Only show scan results, not send to Ralph",
			EnvVar: "RALPH_SCAN_DRY_RUN",
		})
		deviceTemplate = scanCli.String(cli.StringOpt{
			Name:  "device-template",
			Value: "",
			Desc: `Ready to use plugin and components pack to scan typical
                         devices. Available templates: IDRAC, ILO, XEN `,
			EnvVar: "RALPH_SCAN_DEVICE_TEMPLATE",
		})
	)

	scanCli.Action = func() {
		scan.ScanRunner(ip, components, plugins, dryRun, deviceTemplate)
	}

	scanCli.Command("generate-config", "Generate Ralph-Scan config in ~/.ralph-scan/config.yml", GenerateConfig)
	scanCli.Version("v version", "ralph-scan 0.1.0")

	return scanCli

}
