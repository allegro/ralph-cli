package main

import "net"

// Data Center assets: PhysicalHost, VmHost, CloudHost, MesosHost.
// Their json representations should match the shape of Ralph's API endpoint(s)
// associated with them.

// TODO(xor-xor): discuss the relationships between PhysicalHost, VMHost, CloudHost and
// MesosHost (see comments below).

// PhysicalHost represents
type PhysicalHost struct {
	MACAddresses []net.HardwareAddr
	Model        Model  `json:"model,omitempty"`
	SerialNumber string `json:"sn,omitempty"`
	Processors   []Processor
	Memory       []Memory
	Disks        []Disk
	// other fields to be added later
}

// Model represents a model name of a given PhysicalHost (e.g., Dell PowerEdge R620).
type Model struct {
	Name string `json:"name,omitempty"`
}

// VMHost represents a single virtual machine.
// TODO(xor-xor): which one exactly, i.e. what's the difference between CloudHost..?
type VMHost struct{}

// CloudHost represents a single OpenStack node.
// TODO(xor-xor): VM spawned in OpenStack or compute node..?
type CloudHost struct{}

// MesosHost represents a single host being a Mesos node (master or slave).
type MesosHost struct{}

// Hardware components.

// Processor represents a single processor on a given host.
type Processor struct {
	Name  string
	Cores int
	Speed int
}

// Memory represents RAM installed on a given host.
type Memory struct {
	Name  string
	Size  int
	Speed int
}

// Disk represents a single hard drive (be it SSD or "normal" one) on a given host.
type Disk struct {
	Name         string
	Size         int
	SerialNumber string
}

// NetworkCard represents a single network card on a given host.
type NetworkCard struct {
	Model string
	Speed int
	// TODO(xor-xor): consider if storing MAC addresses here would be better, than directly
	// on PhysicalHost.
}

// Other components (well, Software or OperatingSystem are hardly "components" per se),
// to be addressed later.

// Software installed on a given host (PhysicalHost, VMHost, CloudHost, MesosHost).
type Software struct{}

// OperatingSystem present on a given host (PhysicalHost, VMHost, CloudHost, MesosHost).
type OperatingSystem struct{}

// ToPhysicalHost converts ScanResult to PhysicalHost.
func (sr ScanResult) ToPhysicalHost() PhysicalHost {
	return PhysicalHost{}
}

// ToVMHost converts ScanResult to VmHost.
func (sr ScanResult) ToVMHost() VMHost {
	return VMHost{}
}

// ToCloudHost converts ScanResult to CloudHost.
func (sr ScanResult) ToCloudHost() CloudHost {
	return CloudHost{}
}

// ToMesosHost converts ScanResult to MesosHost.
func (sr ScanResult) ToMesosHost() MesosHost {
	return MesosHost{}
}
