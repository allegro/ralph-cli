package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	cli "github.com/jawher/mow.cli"
)

const CONFIG_DIR = ".ralph-scan"
const CONFIG_FILENAME = "config.yml"

const DEFAULT_CONFIG_CONTENT = `global:
  auth:
    username: ralph
    password: ralph
plugins:
  - ILO
  - IDRAC
  - IPMI
`

func getUserHomeDir() string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return dir
}

func getConfigDirFullPath() string {
	userDir := getUserHomeDir()
	configDirPath := path.Join(userDir, CONFIG_DIR)
	return configDirPath
}

func getConfigFilePath() string {
	configDir := getConfigDirFullPath()
	configPath := path.Join(configDir, CONFIG_FILENAME)
	return configPath
}

func pathExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func createConfig(path string) {
	configDirFullPath := getConfigDirFullPath()
	exists := pathExists(configDirFullPath)
	if !exists {
		os.Mkdir(configDirFullPath, 0700)
	}

	defaultConfig := []byte(DEFAULT_CONFIG_CONTENT)
	err := ioutil.WriteFile(path, defaultConfig, 0700)
	if err != nil {
		panic(err)
	}

}

func GenerateConfig(cmd *cli.Cmd) {
	cmd.Action = func() {
		configFilePath := getConfigFilePath()

		configFileExist := pathExists(configFilePath)
		if configFileExist {
			fmt.Printf("config `%s` exists", configFilePath)
		} else {
			fmt.Printf("config `%s` does not exists, creating...", configFilePath)
			createConfig(configFilePath)
		}
	}
}
