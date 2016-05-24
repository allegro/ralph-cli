package main

import (
	"encoding/json"
	"fmt"
	"net"
)

// Data Center assets: PhysicalHost, VmHost, CloudHost, MesosHost.
// Their json representations should match the shape of Ralph's API endpoint(s)
// associated with them.

// PhysicalHost represents information about a physical ("bare metal") host stored
// in Ralph.
type PhysicalHost struct {
	MACAddresses []net.HardwareAddr `json:"mac_addresses,omitempty"`
	Model        Model              `json:"model,omitempty"`
	SerialNumber string             `json:"sn,omitempty"`
	Processors   []Processor        `json:"processors,omitempty"`
	Memory       []Memory           `json:"memory,omitempty"`
	Disks        []Disk             `json:"disks,omitempty"`
	// other fields to be added later
}

// Model represents a model name of a given PhysicalHost (e.g., Dell PowerEdge R620).
type Model struct {
	Name string `json:"model_name"`
}

func (m Model) String() string {
	return fmt.Sprintf("Model{name: %s}", m.Name)
}

// VMHost represents a single virtual machine.
// TODO(xor-xor): Provide some examples.
type VMHost struct{}

// CloudHost represents a single OpenStack VM.
type CloudHost struct{}

// MesosHost represents a single host being a Mesos node (master or slave).
type MesosHost struct{}

// Hardware components.

// Processor represents a single processor on a given host.
type Processor struct {
	Name  string `json:"label"`
	Cores int
	Speed int
}

func (p Processor) String() string {
	return fmt.Sprintf("Processor{name: %s, cores: %d, speed: %d}",
		p.Name, p.Cores, p.Speed)
}

// Memory represents RAM installed on a given host.
type Memory struct {
	Name  string `json:"label"`
	Size  int
	Speed int
}

func (m Memory) String() string {
	return fmt.Sprintf("Memory{name: %s, size: %d, speed: %d}", m.Name, m.Size, m.Speed)
}

// Disk represents a single hard drive (be it SSD or "normal" one) on a given host.
type Disk struct {
	Name         string `json:"model_name"`
	Size         int    `json:"size"`
	SerialNumber string `json:"serial_number"`
}

func (d Disk) String() string {
	return fmt.Sprintf("Disk{model: %s, size: %d, sn: %s}",
		d.Name, d.Size, d.SerialNumber)
}

// MACAddress represents a physical addres of a network card (EthernetComponent).
type MACAddress struct {
	net.HardwareAddr
}

func (m *MACAddress) String() string {
	if len(m.HardwareAddr) == 0 {
		return "(none)"
	}
	return m.HardwareAddr.String()
}

func (m *MACAddress) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(m.String())
	if err != nil {
		return []byte{}, fmt.Errorf("invalid MAC address: %v", err)
	}
	return data, nil
}

func (m *MACAddress) UnmarshalJSON(data []byte) error {
	if string(data) == "\"\"" {
		m.HardwareAddr = []byte{}
		return nil
	}
	var s string
	var msg = "invalid MAC address: %s"
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf(msg, data)
	}
	mac, err := net.ParseMAC(s)
	if err != nil {
		return fmt.Errorf(msg, s)
	}
	m.HardwareAddr = mac
	return nil
}

// EthernetComponent represents a network card on a given host linked to it
// via BaseObject.
type EthernetComponent struct {
	ID         int
	BaseObject BaseObject `json:"base_object"`
	MACAddress MACAddress `json:"mac"`
	Label      string     `json:"label"`
	Speed      string
	Model      string
}

// NewEthernetComponent creates a new EthernetComponent holding info about some network card.
func NewEthernetComponent(mac MACAddress, baseObj *BaseObject, speed string) (*EthernetComponent, error) {
	// TODO(xor-xor): use OPTIONS to get possible values for this field
	if speed == "" {
		speed = "unknown speed"
	}
	return &EthernetComponent{
		BaseObject: *baseObj,
		MACAddress: mac,
		Speed:      speed,
		// Label: label,
		// Model: model
	}, nil
}

func (e *EthernetComponent) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(map[string]interface{}{
		"base_object": e.BaseObject.ID,
		"mac":         e.MACAddress.String(),
		"label":       e.Label,
		"speed":       e.Speed,
		"model":       e.Model,
	})
	if err != nil {
		return []byte{}, fmt.Errorf("error while marshaling EthernetComponent: %v", err)
	}
	return data, nil
}

func (e EthernetComponent) String() string {
	return fmt.Sprintf("EthernetComponent{ID: %d, base object ID: %d, MAC address: %s, Label: %s, Speed: %s, Model: %s}",
		e.ID, e.BaseObject.ID, e.MACAddress, e.Label, e.Speed, e.Model)
}

// ExcludeMgmt filters eths by excluding EthernetComponents associated with given IP
// address, but only when such address is a management one.
// This function should be considered as a temporary solution, and will be removed once
// similar functionality will be implemented in Ralph's API.
func ExcludeMgmt(eths []*EthernetComponent, ip Addr, c *Client) ([]*EthernetComponent, error) {

	type IPAddress struct {
		IsMgmt   bool `json:"is_management"`
		Ethernet *EthernetComponent
	}

	type IPAddressList struct {
		Count   int
		Results []IPAddress
	}

	var ethsFiltered []*EthernetComponent

	q := fmt.Sprintf("address=%s", ip)
	rawBody, err := c.GetFromRalph(APIEndpoints["IPAddress"], q)
	if err != nil {
		return nil, err
	}
	var addrs IPAddressList
	if err := json.Unmarshal(rawBody, &addrs); err != nil {
		return nil, fmt.Errorf("error while unmarshaling IPAddress: %v", err)
	}
	switch {
	case addrs.Count > 1:
		// This shouldn't happen...
		return nil, fmt.Errorf("more than one (%d) record for IPAddress %s", addrs.Count, ip)
	case addrs.Count == 0 || !addrs.Results[0].IsMgmt:
		return eths, nil
	default:
		for _, eth := range eths {
			if eth.ID != addrs.Results[0].Ethernet.ID {
				ethsFiltered = append(ethsFiltered, eth)
			}
		}
	}
	return ethsFiltered, nil
}

// EthernetComponentList represents the shape of data returned by Ralph for EthernetComponent endpoint.
type EthernetComponentList struct {
	Count   int
	Results []EthernetComponent
}

// BaseObject represents an abstract entity used in Ralph as a parent object for
// PhysicalHost, VMHost, CloudHost etc.
type BaseObject struct {
	ID int
	// other fileds to be added later
}

// BaseObjectList represents the shape of data returned by Ralph for the BaseObject
// endpoint.
type BaseObjectList struct {
	Count   int
	Results []BaseObject
}

// GetEthernetComponents fetches EthernetComponents associated with given BaseObject.
func (b *BaseObject) GetEthernetComponents(c *Client) ([]*EthernetComponent, error) {
	q := fmt.Sprintf("base_object=%d", b.ID)
	rawBody, err := c.GetFromRalph(APIEndpoints["EthernetComponent"], q)
	if err != nil {
		return nil, err
	}
	var eths EthernetComponentList
	if err := json.Unmarshal(rawBody, &eths); err != nil {
		return nil, fmt.Errorf("error while unmarshaling EthernetComponent: %v", err)
	}
	ethsPtrs := make([]*EthernetComponent, eths.Count)
	for i := 0; i < eths.Count; i++ {
		ethsPtrs[i] = &eths.Results[i]
	}
	return ethsPtrs, nil
}

// Addr represents and address (IP or FQDN) being scanned.
// TODO(xor-xor): Consider adding IP field here (see my comment in NewAddr).
type Addr string

// NewAddr creates a new Addr from a given string and performs some basic validation on it.
func NewAddr(s string) (Addr, error) {
	// TODO(xor-xor): add support for net.IPNet (via net.ParseCIDR etc.)
	_, err := net.LookupIP(s)
	if err != nil {
		return "", err
	}
	// TODO(xor-xor): Add some other validations/conversions here.
	return Addr(s), nil
}

// GetBaseObject fetches BaseObjects associated with given Addr.
func (a *Addr) GetBaseObject(c *Client) (*BaseObject, error) {
	q := fmt.Sprintf("ip=%s", *a)
	rawBody, err := c.GetFromRalph(APIEndpoints["BaseObject"], q)
	if err != nil {
		return nil, err
	}

	var baseObjs BaseObjectList
	if err := json.Unmarshal(rawBody, &baseObjs); err != nil {
		return nil, fmt.Errorf("error while unmarshaling base object: %v", err)
	}

	switch {
	case baseObjs.Count == 0:
		return nil, fmt.Errorf("IP address %s doesn't have any base objects", *a)
	case baseObjs.Count > 1:
		return nil, fmt.Errorf("IP address %s has more than one base objects", *a)
	default:
		baseObj := baseObjs.Results[0]
		return &baseObj, nil
	}
}

// DiffEthernetComponent represents a set of changes to be made on some EthernetComponents
// associated with a BaseObject (e.g., network cards inserted to a bare-metal host).
// These  changes are grouped in three categories: Create, Update and Delete accounting
// for respective CRUD operations.
type DiffEthernetComponent struct {
	Create []*EthernetComponent
	Update []*EthernetComponent
	Delete []*EthernetComponent
}

// IsEmpty is a helper function for checking if a given DiffEthernetComponent is empty.
func (d *DiffEthernetComponent) IsEmpty() bool {
	if len(d.Create) == 0 && len(d.Update) == 0 && len(d.Delete) == 0 {
		return true
	}
	return false
}

// SendDiffToRalph sends a given DiffEthernetComponent to Ralph. If dryRun is set to true,
// then no changes will be sent to Ralph.
func SendDiffToRalph(c *Client, d *DiffEthernetComponent, dryRun bool) error {
	var msg = "EthernetComponent with MAC address %s %s successfully.\n"
	var err error
	for _, ec := range d.Create {
		data, err := json.Marshal(ec)
		if err != nil {
			return err
		}
		if !dryRun {
			err = c.SendToRalph("POST", APIEndpoints["EthernetComponent"], data)
		}
		if err != nil {
			return err
		}
		fmt.Printf(msg, ec.MACAddress.String(), "created") // TODO(xor-xor): Use logger instead.
	}
	for _, ec := range d.Update {
		data, err := json.Marshal(ec)
		if err != nil {
			return err
		}
		endpoint := fmt.Sprintf("%s/%d", APIEndpoints["EthernetComponent"], ec.ID)
		if !dryRun {
			err = c.SendToRalph("PUT", endpoint, data)
		}
		if err != nil {
			return err
		}
		fmt.Printf(msg, ec.MACAddress.String(), "updated") // TODO(xor-xor): Use logger instead.
	}
	for _, ec := range d.Delete {
		endpoint := fmt.Sprintf("%s/%d", APIEndpoints["EthernetComponent"], ec.ID)
		if !dryRun {
			err = c.SendToRalph("DELETE", endpoint, nil)
		}
		if err != nil {
			return err
		}
		fmt.Printf(msg, ec.MACAddress.String(), "deleted") // TODO(xor-xor): Use logger instead.
	}
	return nil
}

// CompareEthernetComponents compares two sets of EthernetComponents (old and new) and
// creates DiffEthernetComponent holding detected changes.
func CompareEthernetComponents(old, new []*EthernetComponent) (*DiffEthernetComponent, error) {
	var all, create, update, delete []*EthernetComponent
	all = append(all, old...)
	all = append(all, new...)

	for _, eth := range all {
		switch mac := eth.MACAddress; {
		case contains(new, mac) && !contains(old, mac):
			create = append(create, eth)
		case contains(old, mac) && !contains(new, mac):
			delete = append(delete, eth)
		default:
			// We need to find matching eth in old and assign its ID to the new one
			// to make an update. Also, we need to exclude eth equal to ethOld (there's
			// no need to update in such case).
			if eth.ID == 0 {
				var ethOld *EthernetComponent
				for _, ethOld = range old {
					if ethOld.MACAddress.String() == eth.MACAddress.String() {
						eth.ID = ethOld.ID
						break
					}
				}
				if eth.ID == 0 {
					// This really shouldn't happen.
					return nil, fmt.Errorf("error while preparing EthernetComponent with MAC %s for update", eth.MACAddress)
				}
				if !eth.IsEqualTo(ethOld) {
					update = append(update, eth)
				}
			}
		}
	}
	return &DiffEthernetComponent{
		Create: create,
		Delete: delete,
		Update: update,
	}, nil
}

// IsEqualTo compares two EthernetComponents for equality.
func (ec1 *EthernetComponent) IsEqualTo(ec2 *EthernetComponent) bool {
	switch {
	case ec1.BaseObject.ID != ec2.BaseObject.ID:
		return false
	case ec1.MACAddress.String() != ec2.MACAddress.String():
		return false
	case ec1.Label != ec2.Label:
		return false
	case ec1.Speed != ec2.Speed:
		return false
	case ec1.Model != ec2.Model:
		return false
	default:
		return true
	}
}

// Checks if a set of EthernetComponents contains a given MAC address.
func contains(eths []*EthernetComponent, mac MACAddress) bool {
	for _, eth := range eths {
		if eth.MACAddress.String() == mac.String() {
			return true
		}
	}
	return false
}

// Helper function for development/diagnostic purposes.
func printDiffEthernetComponent(diff *DiffEthernetComponent, oldEths, newEths []*EthernetComponent) {
	for _, e := range oldEths {
		fmt.Println("=> old:", e)
	}
	for _, e := range newEths {
		fmt.Println("=> new:", e)
	}
	fmt.Println("===> create:")
	for _, e := range diff.Create {
		fmt.Println("=>", e)
	}
	fmt.Println("===> delete:")
	for _, e := range diff.Delete {
		fmt.Println("=>", e)
	}
	fmt.Println("===> update:")
	for _, e := range diff.Update {
		fmt.Println("=>", e)
	}
}

// Software installed on a given host (PhysicalHost, VMHost, CloudHost, MesosHost).
type Software struct{}

// OperatingSystem present on a given host (PhysicalHost, VMHost, CloudHost, MesosHost).
type OperatingSystem struct{}
