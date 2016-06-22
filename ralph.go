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

// MarshalJSON serializes MACAddress to []byte.
func (m *MACAddress) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(m.String())
	if err != nil {
		return []byte{}, fmt.Errorf("invalid MAC address: %v", err)
	}
	return data, nil
}

// UnmarshalJSON deserializes MACAddress from []byte.
func (m *MACAddress) UnmarshalJSON(data []byte) error {
	// Most management IPs in Ralph won't be associated with any MAC address.
	if string(data) == "\"\"" || string(data) == "null" {
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
func NewEthernetComponent(mac MACAddress, baseObj *BaseObject, speed string) *EthernetComponent {
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
	}
}

// MarshalJSON serializes EthernetComponent to []byte.
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
	var ethsFiltered []*EthernetComponent
	addrs, err := getIPAddresses(fmt.Sprintf("address=%s", ip), c)
	if err != nil {
		return nil, err
	}
	// IP addresses are unique in Ralph, so there's no need to check for addrs.Count > 1.
	if addrs.Count == 0 || !addrs.Results[0].IsMgmt {
		return eths, nil
	}
	for _, eth := range eths {
		if eth.ID != addrs.Results[0].Ethernet.ID {
			ethsFiltered = append(ethsFiltered, eth)
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

// IPAddress is a helper type, i.e. its instances are not meant to be sent to Ralph.
// TODO(xor-xor): Consider merging IPAddress with Addr type.
type IPAddress struct {
	Address      string
	IsMgmt       bool `json:"is_management"`
	ExposeInDHCP bool `json:"dhcp_expose"`
	Ethernet     *EthernetComponent
}

// IPAddressList represents the shape of data returned by Ralph for IPAddress endpoint.
type IPAddressList struct {
	Count   int
	Results []IPAddress
}

// getIPAddress is a helper function for querying "ipaddresses" endpoint.
func getIPAddresses(query string, c *Client) (*IPAddressList, error) {
	var addrs IPAddressList
	rawBody, err := c.GetFromRalph(APIEndpoints["IPAddress"], query)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawBody, &addrs); err != nil {
		return nil, fmt.Errorf("error while unmarshaling IPAddress: %v", err)
	}
	return &addrs, nil
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
// then no changes will be sent to Ralph. If noOutput is set to true, then all the output
// from this function will be silenced (this is mostly for tests).
// Returned statusCodes slice is meant only to facilitate tests, so don't be surprised if you
// see it ignored somewhere in the source code.
func SendDiffToRalph(c *Client, d *DiffEthernetComponent, dryRun bool, noOutput bool) (statusCodes []int, err error) {
	var msg = "EthernetComponent with MAC address %s %s successfully.\n"
	var code int
	for _, ec := range d.Create {
		data, err := json.Marshal(ec)
		if err != nil {
			return statusCodes, err
		}
		if !dryRun {
			code, err = c.SendToRalph("POST", APIEndpoints["EthernetComponent"], data)
			statusCodes = append(statusCodes, code)
		}
		if err != nil {
			return statusCodes, err
		}
		if !noOutput {
			fmt.Printf(msg, ec.MACAddress.String(), "created") // TODO(xor-xor): Use logger instead.
		}
	}
	for _, ec := range d.Update {
		data, err := json.Marshal(ec)
		if err != nil {
			return statusCodes, err
		}
		endpoint := fmt.Sprintf("%s/%d", APIEndpoints["EthernetComponent"], ec.ID)
		if !dryRun {
			code, err = c.SendToRalph("PUT", endpoint, data)
			statusCodes = append(statusCodes, code)
		}
		if err != nil {
			return statusCodes, err
		}
		if !noOutput {
			fmt.Printf(msg, ec.MACAddress.String(), "updated") // TODO(xor-xor): Use logger instead.
		}
	}
	for _, ec := range d.Delete {
		endpoint := fmt.Sprintf("%s/%d", APIEndpoints["EthernetComponent"], ec.ID)
		if !dryRun {
			code, err = c.SendToRalph("DELETE", endpoint, nil)
			statusCodes = append(statusCodes, code)
		}
		if err != nil {
			return statusCodes, err
		}
		if !noOutput {
			fmt.Printf(msg, ec.MACAddress.String(), "deleted") // TODO(xor-xor): Use logger instead.
		}
	}
	return statusCodes, nil
}

// ExcludeExposedInDHCP takes DiffEthernetComponent, and examines EthernetComponents from d.Delete list.
// In (quite unlikely, but possible) case of finding such EthernetComponent, it is excluded from said diff,
// and warning message is printed for user (unless noOutput is set to true, which is meant for testing).
func ExcludeExposedInDHCP(d *DiffEthernetComponent, c *Client, noOutput bool) (*DiffEthernetComponent, error) {
	var ethsFiltered []*EthernetComponent
	for _, ec := range d.Delete {
		ip, err := checkIfExposedInDHCP(&ec.MACAddress, c)
		if err != nil {
			return nil, err
		}
		if ip.Address != "" {
			if !noOutput {
				fmt.Printf("WARNING: EthernetComponent with MAC address %s cannot be deleted, "+
					"because IP address associated with it (%s) is marked as \"exposed in DHCP\" "+
					"in Ralph. Please use a suitable transition from Ralph's GUI for that.\n",
					ec.MACAddress.String(), ip.Address) // TODO(xor-xor): Use logger instead.
			}
			continue
		}
		ethsFiltered = append(ethsFiltered, ec)
	}
	d.Delete = ethsFiltered
	return d, nil
}

// checkIfExposedInDHCP is a helper function for ExcludeExposedInDHCP.
func checkIfExposedInDHCP(m *MACAddress, c *Client) (IPAddress, error) {
	addrs, err := getIPAddresses(fmt.Sprintf("ethernet__mac=%s", m.String()), c)
	if err != nil {
		return IPAddress{}, err
	}
	for _, ip := range addrs.Results {
		if ip.ExposeInDHCP == true {
			return ip, nil
		}
	}
	return IPAddress{}, nil
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
func (e *EthernetComponent) IsEqualTo(ec *EthernetComponent) bool {
	switch {
	case e.BaseObject.ID != ec.BaseObject.ID:
		return false
	case e.MACAddress.String() != ec.MACAddress.String():
		return false
	case e.Label != ec.Label:
		return false
	case e.Speed != ec.Speed:
		return false
	case e.Model != ec.Model:
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

// Software installed on a given host (PhysicalHost, VMHost, CloudHost, MesosHost).
type Software struct{}

// OperatingSystem present on a given host (PhysicalHost, VMHost, CloudHost, MesosHost).
type OperatingSystem struct{}
