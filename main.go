package main

import (
	"github.com/quamilek/ralph-scan/cmd"
	"os"
)

func main() {
	cli := cmd.ScanCli()
	cli.Run(os.Args)
}
