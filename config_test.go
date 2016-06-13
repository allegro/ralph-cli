package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

var configTestFixturesDir = "./config_test_fixtures"

func TestGetConfig(t *testing.T) {
	var cases = map[string]struct {
		fixtureFile string
		want        *Config
		errMsg      string
	}{
		"#0 Everything OK": {
			fixtureFile: "config.toml",
			want: &Config{
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
