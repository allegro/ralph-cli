package main

import "net"

// Data Center assets: PhysicalHost, VmHost, CloudHost, MesosHost.
// Their json representations should match the shape of Ralph's API endpoint(s)
// associated with them.
type PhysicalHost struct {
	MACAddresses []net.HardwareAddr
	Model        Model  `json:"model,omitempty"`
	SerialNumber string `json:"sn,omitempty"`
	Processors   []Processor
	Memory       []Memory
	Disks        []Disk
	// other fields to be added later
}
type Model struct {
	Name string `json:"name,omitempty"`
}
type VmHost struct{}
type CloudHost struct{}
type MesosHost struct{}

// Hardware components.
type Processor struct {
	Name  string
	Cores int
	Speed int
}
type Memory struct {
	Name  string
	Size  int
	Speed int
}
type Disk struct {
	Name         string
	Size         int
	SerialNumber string
}
type NetworkCard struct{}

// Other components, to be addressed later.
type Software struct{}
type OperatingSystem struct{}

// ToPhysicalHost converts ScanResult to PhysicalHost.
func (sr ScanResult) ToPhysicalHost() PhysicalHost {
	return PhysicalHost{}
}

// ToVmHost converts ScanResult to VmHost.
func (sr ScanResult) ToVmHost() VmHost {
	return VmHost{}
}

// To CloudHost converts ScanResult to CloudHost.
func (sr ScanResult) ToCloudHost() CloudHost {
	return CloudHost{}
}

// ToMesosHost converts ScanResult to MesosHost.
func (sr ScanResult) ToMesosHost() MesosHost {
	return MesosHost{}
}
