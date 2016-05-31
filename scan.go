package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Script represents a single, user script which performs the actual scan
// of an address (IP/network or FQDN).
type Script struct {
	Name      string
	LocalPath string
	RepoURL   string
	Manifest  *Manifest
}

// NewScript creates a new instance of Script given as string and performs some basic
// validation of a file associated with it (e.g., is it executable).
func NewScript(fileName string) (Script, error) {
	loc, err := GetCfgDirLocation()
	if err != nil {
		return Script{}, err
	}
	path := filepath.Join(loc, "scripts", fileName)
	finfo, err := os.Stat(path)
	if err != nil {
		return Script{}, err
	}
	exec := finfo.Mode() & 0100
	if exec == 0 {
		return Script{}, fmt.Errorf("file %s is not executable for the owner", path)
	}
	return Script{
		Name:      fileName,
		LocalPath: path,
		RepoURL:   "",
		Manifest:  nil,
	}, nil
}

// Scan performs a scan of a given address (IP or FQDN).
func (s Script) Run(addr Addr) (*ScanResult, error) {
	var res ScanResult
	var err error
	cmd := exec.Command(s.LocalPath, string(addr))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &res, fmt.Errorf("error running script %s: %s\nstderr: %s",
			s.LocalPath, err, cmd.Stderr)
	}
	err = json.Unmarshal(output, &res)
	return &res, err
}

// ScanResult holds parsed output of a scan script.
type ScanResult struct {
	// TODO(xor-xor): Consider adding here a field holding an ADDR being scanned.
	MACAddresses []MACAddress `json:"mac_addresses"`
	Disks        []Disk
	Memory       []Memory
	Model        string `json:"model_name"`
	Processors   []Processor
	SN           string `json:"serial_number"`
}

func (sr ScanResult) String() string {
	return fmt.Sprintf("MACAddresses: %s\nDisks: %s\nMemory: %s\nModel: %s\nProcessors: %s\nSerial Number: %s\n",
		sr.MACAddresses, sr.Disks, sr.Memory, sr.Model, sr.Processors, sr.SN)
}
