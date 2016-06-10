package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the configuration for ralph-cli.
type Config struct {
	Debug                  bool
	LogOutput              string // e.g. logstash
	ClientTimeout          int
	RalphAPIURL            string
	RalphAPIKey            string
	ManagementUserName     string
	ManagementUserPassword string
}

// DefaultCfg provides defaults for Config. Fields with zero-values for their respective
// fields are ommited.
var DefaultCfg = Config{
	ClientTimeout:          10, // seconds
	RalphAPIURL:            "change_me",
	RalphAPIKey:            "change_me",
	ManagementUserName:     "change_me",
	ManagementUserPassword: "change_me",
}

// List of scripts that are bundled with ralph-cli.
var bundledScripts = []string{"idrac.py", "ilo.py"}

// GetCfgDirLocation gets path to current user's home dir and appends ".ralph-cli"
// to it, if baseDir is an empty string, otherwise appends ".ralph-cli" to baseDir path
// (the former case is meant mostly for facilitation of testing).
func GetCfgDirLocation(baseDir string) (string, error) {
	switch {
	case baseDir == "":
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		baseDir = user.HomeDir
	default:
		if _, err := os.Stat(baseDir); err != nil {
			return "", err
		}
	}
	return filepath.Join(baseDir, ".ralph-cli"), nil
}

// PrepareCfgDir creates config dir given as cfgDir (for most cases it will ber
// ~/.ralph-cli). It also creates default config file, and copies bundled scripts
// to the scripts subdir.
func PrepareCfgDir(cfgDir, cfgFileName string) error {
	var err error
	if err = createCfgDir(cfgDir); err != nil {
		return err
	}

	// Create default config file.
	var cfgFile = filepath.Join(cfgDir, cfgFileName)
	if !cfgFileExists(cfgFile) {
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(DefaultCfg); err != nil {
			return err
		}
		if err := ioutil.WriteFile(cfgFile, buf.Bytes(), os.FileMode(0600)); err != nil {
			return err
		}
	}

	// Copy bundled scripts.
	var scriptsDir = filepath.Join(cfgDir, "scripts")
	for _, script := range bundledScripts {
		if _, err := os.Stat(filepath.Join(scriptsDir, script)); os.IsNotExist(err) {
			if err = RestoreAsset(scriptsDir, script); err != nil {
				return err
			}
		}
	}
	return nil
}

// createCfgDir is a helper function for creating ~/.ralph-cli dir (used by PrepareCfgDir).
func createCfgDir(loc string) error {
	var err error
	_, err = os.Stat(loc)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Join(loc, "scripts"), os.FileMode(0755))
		// Add other subdirs to be created here if needed.
	}
	return err
}

// GetConfig loads ralph-cli configuration from cfgDir (in most cases, it will be
// ~/.ralph-cli), performs basic validation on it and supplements some of the missing
// values with their defaults.
func GetConfig(cfgFile string) (*Config, error) {
	cfg, err := readConfig(cfgFile)
	if err != nil {
		return nil, err
	}
	err = cfg.validate()
	if err != nil {
		return nil, err
	}
	cfg.getDefaults()
	return cfg, nil
}

// readConfig reads contents of cfgFile and returns it as *Config. If cfgFile points
// to a non-existing location, then returned config will be populated with settings
// copied from DefaultCfg.
func readConfig(cfgFile string) (*Config, error) {
	var cfg Config
	switch {
	case !cfgFileExists(cfgFile):
		cfg = DefaultCfg
	default:
		if err := checkCfgFilePerms(cfgFile); err != nil {
			return nil, err
		}
		if _, err := toml.DecodeFile(cfgFile, &cfg); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func cfgFileExists(cfgFile string) bool {
	_, err := os.Stat(cfgFile)
	if err != nil {
		return false
	}
	return true
}

// checkCfgFilePerms checks if cfgFile has 0600 permissions and returns an error if not.
// Such permissions are necessary, since cfgFile may contain sensitive information
// (e.g., passwords)
func checkCfgFilePerms(cfgFile string) error {
	// We assume here that cfgFile already exists (use cfgFileExists for such check).
	rwOwnerMask := os.FileMode(0600)
	finfo, _ := os.Stat(cfgFile)
	if finfo.Mode()&^rwOwnerMask != 0 {
		return fmt.Errorf("config file %q should only have read+write permissions for its owner", cfgFile)
	}
	return nil
}

// validate performs some sanity checks/normalizations on Config.
func (c *Config) validate() error {
	if c.RalphAPIKey == "" {
		return fmt.Errorf("config error: Ralph API key is missing")
	}
	if c.RalphAPIURL == "" {
		return fmt.Errorf("config error: Ralph API URL is missing")
	}
	// TODO(xor-xor): Investigate why url.Parse happily accepts stuff like "httplocalhost" or
	// "http/localhost/api", and add some additional checks here for such cases.
	// TODO(xor-xor): Get rid of Query/Fragment if present in URL.
	u, err := url.Parse(c.RalphAPIURL)
	if err != nil {
		return fmt.Errorf("config error: error while parsing Ralph API URL: %v", err)
	}
	c.RalphAPIURL = u.String()
	return nil
}

// getDefaults supplements missing values for Config fields with their defaults.
func (c *Config) getDefaults() {
	// Unfortunately, there's no easy way to iterate over struct fields, hence we need
	// to enumerate default settings manually here.
	switch {
	case c.ClientTimeout == 0:
		c.ClientTimeout = DefaultCfg.ClientTimeout
	}
}

// CreatePythonVenv creates a virtualenv for Python scripts in ~/.ralph-cli dir.
func CreatePythonVenv() error {
	// TODO(xor-xor): Implement this.
	return nil
}
