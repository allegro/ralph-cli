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
		oldEnv []string
		config *Config
		want   []string
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
			[]string{"GO_WANT_HELPER_PROCESS=1", "MANAGEMENT_USER_NAME=some_user", "MANAGEMENT_USER_PASSWORD=some_password"},
		},
		"#1 Existing management user/pass should be overwritten": {
			[]string{"MANAGEMENT_USER_NAME=old_user", "MANAGEMENT_USER_PASSWORD=old_password"},
			&Config{
				ClientTimeout:          10,
				RalphAPIURL:            "http://localhost:8080/api",
				RalphAPIKey:            "abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				ManagementUserName:     "some_user",
				ManagementUserPassword: "some_password",
			},
			[]string{"MANAGEMENT_USER_NAME=some_user", "MANAGEMENT_USER_PASSWORD=some_password"},
		},
	}
	for tn, tc := range cases {
		got := prepareEnv(tc.oldEnv, tc.config)
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
	script := Script{
		Name:      "idrac.py",
		LocalPath: "/path/to/homedir/.ralph-cli/scripts/idrac.py",
		RepoURL:   "",
		Manifest:  nil,
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

	got, err := script.Run("10.20.30.40", config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if eq, err := checkers.DeepEqual(got, want); !eq {
		t.Errorf("\n%s", err)
	}
}

func TestNewScript(t *testing.T) {
	cfgDir, baseDir, err := GetTempCfgDir()
	defer os.RemoveAll(baseDir)

	want := Script{
		Name:      "idrac.py",
		LocalPath: filepath.Join(cfgDir, "scripts", "idrac.py"),
		RepoURL:   "",
		Manifest:  nil,
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
