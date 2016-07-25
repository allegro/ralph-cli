package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Script represents a single, user script which performs the actual scan of a single
// IP address. Each such script may be associated with a Manifest holding extra
// information needed for launching such script (see descriptions of Manifest struct
// and GetManifest function for more details).
type Script struct {
	Path     string
	Manifest *Manifest
}

var execCommand = exec.Command

// NewScript creates a new instance of Script given as sName. It also loads
// manifest file for this script, if present.
// Scripts should be located in "scripts" subdir of cfgDir. When cfgDir is given as an
// empty string, then "~/.ralph-cli/scripts" will be searched (this is the default; the
// former case is meant mostly for tests).
func NewScript(sName, cfgDir string) (Script, error) {
	sPath := filepath.Join(cfgDir, "scripts", sName)
	mfName := changeExt(sName, "toml")
	mfPath := filepath.Join(cfgDir, "scripts", mfName)
	mf, err := GetManifest(mfPath)
	if err != nil {
		return Script{}, err
	}

	return Script{
		Path:     sPath,
		Manifest: mf,
	}, nil
}

// changeExt replaces the last dot-separated component of fName with newExt, so for
// "idrac.py" and "toml" we will get "idrac.toml", but for "another.idrac.py" and "toml"
// we will get "another.idrac.toml", not "another.toml".
func changeExt(fName, newExt string) string {
	components := strings.Split(fName, ".")
	components[len(components)-1] = newExt
	return strings.Join(components, ".")
}

// Run launches a scan Script on a given address (at this moment, only IPs are fully
// supported). If Script has a Manifest, and the Language in this Manifest is set to
// "python", then the interpreter from a virtualenv associated with this script will
// be used to launch it.
func (s Script) Run(addrToScan Addr, cfg *Config) (*ScanResult, error) {
	var cmd *exec.Cmd
	var err error
	var res ScanResult

	switch {
	case s.Manifest != nil && s.Manifest.Language == "python":
		python := filepath.Join(MakeVenvPath(s), "bin", "python")
		cmd = execCommand(python, s.Path)
	default:
		cmd = execCommand(s.Path)
	}

	// This condition will be false only during some tests (see GetHelperCommand),
	// and in such case, we need to preserve cmd.Env contents, hence this check
	// (i.e., prepareEnv should only be launched when cmd.Env is empty).
	if len(cmd.Env) == 0 {
		cmd.Env = prepareEnv(os.Environ(), addrToScan, cfg)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running script %s: %s\noutput from script:\n-->\n%s<--",
			s.Path, err, string(output))
	}
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling script output: %s", err)
	}
	return &res, nil
}

// prepareEnv is a helper function for Script.Run. It modifies the environment that
// should be used for executing given Script.
func prepareEnv(oldEnv []string, addrToScan Addr, cfg *Config) (newEnv []string) {
	for _, e := range oldEnv {
		pair := strings.Split(e, "=")
		switch {
		case pair[0] == "MANAGEMENT_USER_NAME":
			continue
		case pair[0] == "MANAGEMENT_USER_PASSWORD":
			continue
		case pair[0] == "IP_TO_SCAN":
			continue
		default:
			newEnv = append(newEnv, e)
		}
	}
	newEnv = append(newEnv, fmt.Sprintf("MANAGEMENT_USER_NAME=%s", cfg.ManagementUserName))
	newEnv = append(newEnv, fmt.Sprintf("MANAGEMENT_USER_PASSWORD=%s", cfg.ManagementUserPassword))
	newEnv = append(newEnv, fmt.Sprintf("IP_TO_SCAN=%s", addrToScan))
	return newEnv
}

// ScanResult holds parsed output of a scan script.
type ScanResult struct {
	// TODO(xor-xor): Consider adding here a field holding an ADDR being scanned.
	Ethernets         []Ethernet         `json:"ethernets"`
	Memory            []Memory           `json:"memory"`
	FibreChannelCards []FibreChannelCard `json:"fibre_channel_cards"`
	Disks             []Disk             `json:"disks"`
	Processors        []Processor        `json:"processors"`
	SN                string             `json:"serial_number"`
	FirmwareVersion   string             `json:"firmware_version"`
	BIOSVersion       string             `json:"bios_version"`
	ModelName         string             `json:"model_name"`
}

func (sr ScanResult) String() string {
	return fmt.Sprintf("Ethernets: %s\n\nMemory: %s\n\nFibreChannelCards: %s\n\nDisks: %s\n\nModelName: %s\n\nProcessors: %s\n\nFirmware Version: %s\n\nBIOS Version: %s\n\nSerial Number: %s\n",
		sr.Ethernets, sr.Memory, sr.FibreChannelCards, sr.Disks, sr.ModelName, sr.Processors, sr.FirmwareVersion, sr.BIOSVersion, sr.SN)
}
