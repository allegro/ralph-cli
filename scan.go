package main

import (
	"fmt"
	"net"
)

// Script represents a single, user script which performs the actual scan
// of an IP or network.
type Script struct {
	Name      string
	LocalPath string
	RepoURL   string
	Manifest  *Manifest
}

// ScanResult holds validated output of a scan script.
type ScanResult string

// IP represents a single IP address (it's a wrapper around net.IP).
type IP net.IP

// IPNet represents IP address of a network (it's a wrapper around net.IPNet).
type IPNet net.IPNet

// ScanObject implements scanning of IP addresses or networks.
type ScanObject interface {
	Scan() (ScanResult, error)
}

// Run launches a given Script and return its output as ScanResult.
func (s Script) Run() (ScanResult, error) {
	output := ScanResult("dummy output")
	err := output.validate()
	if err != nil {
		return "", err
	}

	return output, nil
}

// validates ScanResult against some schema (to be added later).
func (sr ScanResult) validate() error {
	return nil
}

// NewIP creates a new instance of IP address.
func NewIP(s string) IP {
	return IP(net.ParseIP(s))
}

// NewIPNet creates a new instance of IPNet (IP address of a network).
func NewIPNet(s string) (*IPNet, error) {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	NewIPNet := IPNet(*ipnet)
	return &NewIPNet, nil
}

// Scan performs a scan of a given IP address.
func (ip IP) Scan() (ScanResult, error) {
	return ScanResult("I'm just a dummy scan result."), nil
}

// Scan performs a scan of a given network.
func (ipnet IPNet) Scan() (ScanResult, error) {
	return ScanResult("I'm just a dummy scan result."), nil
}

// XXX Only for demonstration.
func PerformDummyScan(s *string) {
	ip := NewIP(*s)
	result, _ := ip.Scan()
	fmt.Printf("%s\n", result)
}
