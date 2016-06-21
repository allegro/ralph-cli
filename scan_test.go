package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/juju/testing/checkers"
)

var scanTestFixturesDir = "./scan_test_fixtures"

// This is not a real test. It is used as a helper process for TestRun (see go doc for GetHelperCommand).
func TestRunHelperProcess(t *testing.T) {
	var fixtureFile = "script_output.json"
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	output, err := LoadFixture(scanTestFixturesDir, fixtureFile)
	if err != nil {
		t.Fatalf("file: %s\n%s", fixtureFile, err)
	}
	fmt.Fprintf(os.Stdout, output)
	os.Exit(0)
}

func TestPrepareEnv(t *testing.T) {
	var cases = map[string]struct {
		oldEnv     []string
		config     *Config
		addrToScan Addr
		want       []string
	}{
		"#0 Existing Cmd.Env shouldn't be destroyed": {
			[]string{"GO_WANT_HELPER_PROCESS=1"},
			&Config{
				ClientTimeout:          10,
				RalphAPIURL:            "http://localhost:8080/api",
				RalphAPIKey:            "abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				ManagementUserName:     "some_user",
				ManagementUserPassword: "some_password",
			},
			Addr("10.20.30.40"),
			[]string{"GO_WANT_HELPER_PROCESS=1", "MANAGEMENT_USER_NAME=some_user", "MANAGEMENT_USER_PASSWORD=some_password", "IP_TO_SCAN=10.20.30.40"},
		},
		"#1 Existing management user/pass/IP should be overwritten": {
			[]string{"MANAGEMENT_USER_NAME=old_user", "MANAGEMENT_USER_PASSWORD=old_password", "IP_TO_SCAN=11.22.33.44"},
			&Config{
				ClientTimeout:          10,
				RalphAPIURL:            "http://localhost:8080/api",
				RalphAPIKey:            "abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				ManagementUserName:     "some_user",
				ManagementUserPassword: "some_password",
			},
			Addr("10.20.30.40"),
			[]string{"MANAGEMENT_USER_NAME=some_user", "MANAGEMENT_USER_PASSWORD=some_password", "IP_TO_SCAN=10.20.30.40"},
		},
	}
	for tn, tc := range cases {
		got := prepareEnv(tc.oldEnv, tc.addrToScan, tc.config)
		if !TestEqStr(got, tc.want) {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}

func TestRun(t *testing.T) {
	execCommand = GetHelperCommand("TestRunHelperProcess")
	defer func() { execCommand = exec.Command }()

	config := &Config{
		ClientTimeout:          10,
		RalphAPIURL:            "http://localhost:8080/api",
		RalphAPIKey:            "abcdefghijklmnopqrstuwxyz0123456789ABCDE",
		ManagementUserName:     "some_user",
		ManagementUserPassword: "some_password",
	}
	want := &ScanResult{
		MACAddresses: []MACAddress{
			macs["aa:aa:aa:aa:aa:aa"],
			macs["aa:bb:cc:dd:ee:ff"],
			macs["a1:b2:c3:d4:e5:f6"],
			macs["74:86:7a:ee:20:e8"],
		},
		Disks: []Disk{
			Disk{Name: "ATA Samsung SSD 840", Size: 476, SerialNumber: "S1AXNSAD8000000"},
			Disk{Name: "ATA Samsung SSD 840", Size: 476, SerialNumber: "S1AXNSAD8000001"},
		},
		Memory: []Memory{
			Memory{Name: "Samsung DDR3 DIMM", Size: 16384, Speed: 1600},
			Memory{Name: "Samsung DDR3 DIMM", Size: 16384, Speed: 1600},
			Memory{Name: "Samsung DDR3 DIMM", Size: 16384, Speed: 1600},
			Memory{Name: "Samsung DDR3 DIMM", Size: 16384, Speed: 1600},
		},
		Model: "Dell PowerEdge R620",
		Processors: []Processor{
			Processor{Name: "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz", Cores: 8, Speed: 3600},
			Processor{Name: "Intel(R) Xeon(R) CPU E5-2650 v2 @ 2.60GHz", Cores: 8, Speed: 3600},
		},
		SN: "UUUZZZ1",
	}

	var cases = map[string]struct {
		addrToScan Addr
		config     *Config
		script     Script
		want       *ScanResult
	}{
		"#0 Python script with manifest": {
			addrToScan: Addr("10.20.30.40"),
			config:     config,
			script: Script{
				Path:     "/path/to/homedir/.ralph-cli/scripts/script_with_manifest.py",
				Manifest: &Manifest{Language: "python"},
			},
			want: want,
		},
		"#1 Python script without manifest": {
			addrToScan: Addr("10.20.30.40"),
			config:     config,
			script: Script{
				Path:     "/path/to/homedir/.ralph-cli/scripts/script_without_manifest.py",
				Manifest: nil,
			},
			want: want,
		},
	}
	for tn, tc := range cases {
		got, err := tc.script.Run(tc.addrToScan, tc.config)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestNewScript(t *testing.T) {
	cfgDir, baseDir, err := GetTempCfgDir()
	defer os.RemoveAll(baseDir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	want := Script{
		Path: filepath.Join(cfgDir, "scripts", "idrac.py"),
		Manifest: &Manifest{
			Path:            filepath.Join(cfgDir, "scripts", "idrac.toml"),
			Language:        "python",
			LanguageVersion: 3,
			Requirements: []requirement{
				{Name: "requests", Version: "2.10.0"},
			},
		},
	}

	got, err := NewScript("idrac.py", cfgDir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if eq, err := checkers.DeepEqual(got, want); !eq {
		t.Errorf("\n%s", err)
	}
}

// TODO(xor-xor): Add test cases for missing and non-executable script.
