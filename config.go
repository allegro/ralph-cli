package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

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

// List of files (scripts, manifests) that are bundled with ralph-cli.
var bundledFiles = []string{
	"idrac.py",
	"idrac.toml",
	"ilo.py",
	"ilo.toml",
}

// GetCfgDirLocation generates the path to the config dir. It gets the path to current
// user's home dir and if baseDir is an empty string, it appends ".ralph-cli" to it,
// otherwise appends ".ralph-cli" to baseDir path (the former case is meant mostly for
// facilitation of testing).
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
	if !fileExists(cfgFile) {
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(DefaultCfg); err != nil {
			return err
		}
		if err := ioutil.WriteFile(cfgFile, buf.Bytes(), os.FileMode(0600)); err != nil {
			return err
		}
	}

	// Copy bundled files (scripts and manifests).
	var scriptsDir = filepath.Join(cfgDir, "scripts")
	for _, file := range bundledFiles {
		if _, err := os.Stat(filepath.Join(scriptsDir, file)); os.IsNotExist(err) {
			if err = RestoreAsset(scriptsDir, file); err != nil {
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
	case !fileExists(cfgFile):
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

func fileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}
	return true
}

// checkCfgFilePerms checks if cfgFile has 0600 permissions and returns an error if not.
// Such permissions are necessary, since cfgFile may contain sensitive information
// (e.g., passwords)
func checkCfgFilePerms(cfgFile string) error {
	// We assume here that cfgFile already exists (use fileExists for such check).
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

// Manifest represents the contents of a .toml file holding additional information,  which may be
// helpful/required to run user's script (e.g., language, version, requirements etc.).
type Manifest struct {
	Path            string `toml:"-"`
	Language        string
	LanguageVersion int
	Requirements    []requirement `toml:"requirement"`
}

// requirement is a helper type for Manifest. It shouldn't be used alone/separately.
type requirement struct {
	Name    string
	Version string
}

// GetManifest loads manifest file pointed by path argument. Each manifest file should have the same
// name as the script file associated with it (e.g., having a script "idrac.py", its manifest file
// should be named "idrac.toml").
// At this moment manifest files are not required, although it may change in the future (especially
// when we add some more information to them).
func GetManifest(path string) (*Manifest, error) {
	var mf Manifest
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	if _, err := toml.DecodeFile(path, &mf); err != nil {
		return nil, fmt.Errorf("error reading manifest file %s: %v", path, err)
	}
	mf.Path = path
	mf.Language = strings.ToLower(mf.Language)
	if err := mf.validate(); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &mf, nil
}

// validate performs some sanity checks and normalizations on Manifest.
func (m *Manifest) validate() error {
	switch {
	case m.Language == "":
		return fmt.Errorf("manifest error: Language field is missing in %s", m.Path)
	case m.Language == "python" && (m.LanguageVersion != 2 && m.LanguageVersion != 3):
		return fmt.Errorf("manifest error: LanguageVersion field for Python should be either 2 or 3")
	}
	for _, r := range m.Requirements {
		switch {
		case r.Name == "":
			return fmt.Errorf("manifest error: requirement with empty name field in %s", m.Path)
		}
	}
	return nil
}

// VenvExists returns true if there's a virtualenv for a given Python script, or false otherwise.
// We can add more sophisticated heuristics here, but at this moment, checking for bin/activate
// script should be enough.
func VenvExists(s Script) bool {
	venvPath := MakeVenvPath(s)
	file := filepath.Join(venvPath, "bin", "activate")
	return fileExists(file)
}

// MakeVenvPath generates the absolute path to virtualenv associated with the given script by
// replacing from its path the last dot-separated component with "_env" suffix (e.g., for
// "/home/user/.ralph-cli/scripts/idrac.py" we will get "/home/user/.ralph-cli/scripts/idrac_env").
func MakeVenvPath(s Script) string {
	scriptsDir := filepath.Dir(s.Path)
	baseName := strings.TrimSuffix(filepath.Base(s.Path), ".py")
	venvName := strings.Join([]string{baseName, "env"}, "_")
	return filepath.Join(scriptsDir, venvName)
}

// CreatePythonVenv creates a virtualenv for Python scripts in ~/.ralph-cli/scripts and returns
// its path.
func CreatePythonVenv(s Script) (venvPath string, err error) {
	var python string
	venvPath = MakeVenvPath(s)
	switch {
	case s.Manifest.LanguageVersion == 3:
		python = "python3"
	case s.Manifest.LanguageVersion == 2:
		python = "python2"
	default:
		python = "python"
	}
	cmd := execCommand("virtualenv", "-p", python, venvPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error creating python virtualenv: %v", err)
	}
	return venvPath, nil
}

// InstallPythonReqs takes a list of Requirements from Manifest associated with a given Script,
// and installs them with "pip install" into a virtualenv pointed by venvPath.
func InstallPythonReqs(venvPath string, s Script) error {
	numReqs := len(s.Manifest.Requirements)
	if numReqs > 0 {
		var req string
		var args = make([]string, numReqs+1)
		args[0] = "install"
		for i, r := range s.Manifest.Requirements {
			switch {
			case r.Version == "":
				req = r.Name
			default:
				req = strings.Join([]string{r.Name, r.Version}, "==")
			}
			args[i+1] = req
		}
		pip := filepath.Join(venvPath, "bin", "pip")
		cmd := execCommand(pip, args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("error installing requirements for %s: %v; output from pip:\n-->\n%s<--",
				s.Path, err, string(output))
		}
	}
	return nil
}
