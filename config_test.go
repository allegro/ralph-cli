package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/juju/testing/checkers"
)

var configTestFixturesDir = "./config_test_fixtures"

// Unfortunately, git doesn't store file permissions, so we need to restore them on
// files with fixtures before running tests from this file.
func init() {
	var perms = map[string]os.FileMode{
		"config.toml":                   0600,
		"config_api_key_missing.toml":   0600,
		"config_api_url_missing.toml":   0600,
		"config_missing_fields.toml":    0600,
		"config_wrong_permissions.toml": 0666,
	}
	for fileName, mode := range perms {
		err := os.Chmod(filepath.Join(configTestFixturesDir, fileName), mode)
		if err != nil {
			// TODO(xor-xor): Wire up logger here.
			fmt.Printf("err: %s", err)
			os.Exit(1)
		}
	}
}

// Helper process for TestCreatePythonVenv.
func TestDummyHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

func TestGetConfig(t *testing.T) {
	var cases = map[string]struct {
		fixtureFile string
		want        *Config
		errMsg      string
	}{
		"#0 Everything OK": {
			fixtureFile: "config.toml",
			want: &Config{
				Path:                   filepath.Join(configTestFixturesDir, "config.toml"),
				Debug:                  false,
				LogOutput:              "",
				ClientTimeout:          10,
				RalphAPIURL:            "http://localhost:8080/api",
				RalphAPIKey:            "abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				ManagementUserName:     "some_user",
				ManagementUserPassword: "some_password",
			},
			errMsg: "",
		},
		"#1 Wrong permissions": {
			fixtureFile: "config_wrong_permissions.toml",
			want:        nil,
			errMsg:      "should only have read+write permissions for its owner",
		},
		"#2 Ralph API key is missing": {
			fixtureFile: "config_api_key_missing.toml",
			want:        nil,
			errMsg:      "Ralph API key is missing",
		},
		"#3 Ralph API URL is missing": {
			fixtureFile: "config_api_url_missing.toml",
			want:        nil,
			errMsg:      "Ralph API URL is missing",
		},
		"#4 When config file is missing, DefaultCfg should be used": {
			fixtureFile: "does_not_exist.toml",
			want: &Config{
				Debug:                  false,
				LogOutput:              "",
				ClientTimeout:          10,
				RalphAPIURL:            "change_me",
				RalphAPIKey:            "change_me",
				ManagementUserName:     "change_me",
				ManagementUserPassword: "change_me",
			},
			errMsg: "",
		},
		"#5 Some missing fields (e.g. ClientTimeout) are supplemented from DefaultCfg": {
			fixtureFile: "config_missing_fields.toml",
			want: &Config{
				Path:                   filepath.Join(configTestFixturesDir, "config_missing_fields.toml"),
				Debug:                  false,
				LogOutput:              "",
				ClientTimeout:          10,
				RalphAPIURL:            "http://localhost:8080/api",
				RalphAPIKey:            "abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				ManagementUserName:     "some_user",
				ManagementUserPassword: "some_password",
			},
			errMsg: "",
		},
	}
	for tn, tc := range cases {
		cfgFile := filepath.Join(configTestFixturesDir, tc.fixtureFile)
		got, err := GetConfig(cfgFile)
		switch {
		case tc.errMsg != "":
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("%s\ndidn't get expected string: %q in err msg: %q", tn, tc.errMsg, err)
			}
		default:
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if *got != *tc.want {
				t.Errorf("%s\n got: %+v\nwant: %+v", tn, *got, *tc.want)
			}
		}
	}
}

func TestGetCfgDirLocation(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	cfgDir := filepath.Join(user.HomeDir, ".ralph-cli")

	var cases = map[string]struct {
		baseDir string
		want    string
	}{
		"#0 User's home dir should be used as baseDir by default": {
			"",
			cfgDir,
		},
	}
	for tn, tc := range cases {
		got, err := GetCfgDirLocation(tc.baseDir)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}

	}

}

func TestGetManifest(t *testing.T) {
	cfgDir, baseDir, err := GetTempCfgDir()
	defer os.RemoveAll(baseDir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	var cases = map[string]struct {
		path string
		want *Manifest
	}{
		"#0 Manifest file exists": {
			path: filepath.Join(cfgDir, "scripts", "idrac.toml"),
			want: &Manifest{
				Path:            filepath.Join(cfgDir, "scripts", "idrac.toml"),
				Language:        "python",
				LanguageVersion: 3,
				Requirements: []requirement{
					{Name: "requests", Version: "2.10.0"},
				},
			},
		},
		"#1 Manifest file does not exist": {
			path: filepath.Join(cfgDir, "scripts", "does_not_exist.toml"),
			want: nil,
		},
	}

	for tn, tc := range cases {
		got, err := GetManifest(tc.path)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestValidateManifest(t *testing.T) {
	var manifestPath = "/home/user/.ralph-cli/scripts/manifest.toml"
	var cases = map[string]struct {
		manifest *Manifest
		want     error
		errMsg   string
	}{
		"#0 Valid manifest": {
			manifest: &Manifest{
				Path:            manifestPath,
				Language:        "python",
				LanguageVersion: 3,
				Requirements: []requirement{
					{Name: "requests", Version: "2.10.0"},
					{Name: "python-hpilo", Version: "3.8"},
				},
			},
			want: nil,
		},
		"#1 Missing Language field": {
			manifest: &Manifest{
				Path:            manifestPath,
				LanguageVersion: 3,
				Requirements: []requirement{
					{Name: "requests", Version: "2.10.0"},
					{Name: "python-hpilo", Version: "3.8"},
				},
			},
			want: fmt.Errorf("validation error in %s: Language field is missing", manifestPath),
		},
		"#2 Invalid LanguageVersion for Python": {
			manifest: &Manifest{
				Path:            manifestPath,
				Language:        "python",
				LanguageVersion: 4,
				Requirements: []requirement{
					{Name: "requests", Version: "2.10.0"},
					{Name: "python-hpilo", Version: "3.8"},
				},
			},
			want: fmt.Errorf("validation error in %s: LanguageVersion field for Python should be either 2 or 3", manifestPath),
		},
		"#3 Requirement with empty Name field": {
			manifest: &Manifest{
				Path:            manifestPath,
				Language:        "python",
				LanguageVersion: 3,
				Requirements: []requirement{
					{Name: "requests", Version: "2.10.0"},
					{Name: "", Version: "3.8"},
				},
			},
			want: fmt.Errorf("validation error in %s: unknown requirement (empty name field)", manifestPath),
		},
	}

	for tn, tc := range cases {
		got := tc.manifest.validate()
		switch {
		case tc.want == nil:
			if got != tc.want {
				t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
			}
		default:
			if got.Error() != tc.want.Error() {
				t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
			}
		}
	}
}

func TestVenvExists(t *testing.T) {
	cfgDir, baseDir, err := GetTempCfgDir()
	defer os.RemoveAll(baseDir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	venvBinPath := filepath.Join(cfgDir, "scripts", "idrac_env", "bin")
	err = os.MkdirAll(venvBinPath, os.FileMode(0755))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	activateFile, err := os.Create(filepath.Join(venvBinPath, "activate"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer activateFile.Close()

	var cases = map[string]struct {
		script Script
		want   bool
	}{
		"#0 Virtualenv exists": {
			script: Script{
				Path: filepath.Join(cfgDir, "scripts", "idrac.py"),
			},
			want: true,
		},
		"#1 Virtualenv does not exist": {
			script: Script{
				Path: filepath.Join(cfgDir, "scripts", "does_not_exist.py"),
			},
			want: false,
		},
	}

	for tn, tc := range cases {
		got := VenvExists(tc.script)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}

// TODO(xor-xor): Improve TestCreatePythonVenv by modifying helper process in a way that
// enables  checking the args passed to execCommand (maybe by using some channel available
// via  global variable..?).
// TODO(xor-xor): Create a test for InstallPythonReqs when the aforementioned modification
// will be ready (it doesn't make sense to test this function now).
func TestCreatePythonVenv(t *testing.T) {
	execCommand = GetHelperCommand("TestDummyHelperProcess")
	defer func() { execCommand = exec.Command }()

	scriptsPath := "/user/home/.ralph-cli/scripts"
	script := Script{
		Path: filepath.Join(scriptsPath, "idrac.py"),
		Manifest: &Manifest{
			Path:            filepath.Join(scriptsPath, "idrac.toml"),
			Language:        "python",
			LanguageVersion: 3,
		},
	}
	want := filepath.Join(scriptsPath, "idrac_env")

	got, err := CreatePythonVenv(script)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if got != want {
		t.Errorf("\n got: %v\nwant: %v", got, want)
	}
}
