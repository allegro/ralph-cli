package main

import (
	"os"
	"os/user"
	"path/filepath"
)

// Config holds the configuration for ralph-cli.
type Config struct {
	Debug     bool
	LogOutput string // e.g. logstash
}

// List of scripts that are bundled with ralph-cli (at this moment, only idrac.py).
var bundledScripts = []string{"idrac.py"}

// GetCfgDirLocation gets path to current user's home dir and appends ".ralph-cli" to it.
func GetCfgDirLocation() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(user.HomeDir, ".ralph-cli"), nil
}

// PrepareCfgDir creates ~/.ralph-cli dir with its subdirs and copies bundled
// scripts to "scripts" dir.
func PrepareCfgDir() error {
	var err error
	loc, err := GetCfgDirLocation()
	if err != nil {
		return err
	}
	err = createCfgDir(loc)
	if err != nil {
		return err
	}
	var scriptsDir = filepath.Join(loc, "scripts")
	// Copy bundled scripts.
	for _, script := range bundledScripts {
		if _, err := os.Stat(filepath.Join(scriptsDir, script)); os.IsNotExist(err) {
			err = RestoreAsset(scriptsDir, script)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// createCfgDir is a helper function for creating ~/.ralph-cli dir (used by
// PrepareCfgDir).
func createCfgDir(loc string) error {
	var err error
	_, err = os.Stat(loc)
	if os.IsNotExist(err) {
		mode := os.FileMode(int(0755))
		err = os.MkdirAll(filepath.Join(loc, "scripts"), mode)
		// Add other subdirs to be created here if needed.
	}
	return err
}

// GetConfig loads ralph-cli configuration from ~/.ralph-cli dir.
func GetConfig() (Config, error) {
	// TODO(xor-xor): Implement this.
	return Config{
		Debug:     false,
		LogOutput: "",
	}, nil
}

// CreateDefaultCfg creates default ralph-cli config in ~/.ralph-cli dir,
// if not present.
func CreateDefaultCfg() error {
	// TODO(xor-xor): Implement this.
	return nil
}

// CreatePythonVenv creates a virtualenv for Python scripts in ~/.ralph-cli dir.
func CreatePythonVenv() error {
	// TODO(xor-xor): Implement this.
	return nil
}
