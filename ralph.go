package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

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
func (a Addr) GetBaseObject(c *Client) (*BaseObject, error) {
	q := fmt.Sprintf("ip=%s", a)
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
		return nil, fmt.Errorf("IP address %s doesn't have any base objects", a)
	case baseObjs.Count > 1:
		return nil, fmt.Errorf("IP address %s has more than one base objects", a)
	default:
		baseObj := baseObjs.Results[0]
		return &baseObj, nil
	}
}

// BaseObjectList represents the shape of data returned by Ralph for the BaseObject
// endpoint.
type BaseObjectList struct {
	Count   int
	Results []BaseObject
}

// BaseObject represents an abstract entity used in Ralph as a parent object for
// physical hosts and therefore - all components associated with them.
type BaseObject struct {
	ID int
}

// MarshalJSON serializes BaseObject. This method is necessary (in contrast to
// deserialization), because Ralph returns BaseObjects as a nested entity in a
// given component (e.g. Ethernet.BaseObject), but it requires BaseObject.ID as
// a field in JSON being POST-ed (i.e. only ID, not the whole object), so we
// have to "flatten" it here.
func (b *BaseObject) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(b.ID)
	if err != nil {
		return []byte{}, fmt.Errorf("error marshaling BaseObject: %v", err)
	}
	return data, nil
}

// GetEthernets fetches Ethernet objects associated with given BaseObject.
func (b BaseObject) GetEthernets(c *Client) ([]*Ethernet, error) {
	q := fmt.Sprintf("base_object=%d", b.ID)
	rawBody, err := c.GetFromRalph(APIEndpoints["Ethernet"], q)
	if err != nil {
		return nil, err
	}
	var eths EthernetList
	if err := json.Unmarshal(rawBody, &eths); err != nil {
		return nil, fmt.Errorf("error while unmarshaling Ethernet: %v", err)
	}
	ethsPtrs := make([]*Ethernet, eths.Count)
	for i := 0; i < eths.Count; i++ {
		ethsPtrs[i] = &eths.Results[i]
	}
	return ethsPtrs, nil
}

// GetMemory fetches Memory objects associated with given BaseObject.
func (b BaseObject) GetMemory(c *Client) ([]*Memory, error) {
	q := fmt.Sprintf("base_object=%d", b.ID)
	rawBody, err := c.GetFromRalph(APIEndpoints["Memory"], q)
	if err != nil {
		return nil, err
	}
	var mems MemoryList
	if err := json.Unmarshal(rawBody, &mems); err != nil {
		return nil, fmt.Errorf("error while unmarshaling Memory: %v", err)
	}
	memsPtrs := make([]*Memory, mems.Count)
	for i := 0; i < mems.Count; i++ {
		memsPtrs[i] = &mems.Results[i]
	}
	return memsPtrs, nil
}

// MACAddress represents a physical address of a network card (Ethernet).
type MACAddress struct {
	net.HardwareAddr
}

func (m MACAddress) String() string {
	if len(m.HardwareAddr) == 0 {
		return "(none)"
	}
	return m.HardwareAddr.String()
}

// MarshalJSON serializes MACAddress to []byte.
func (m MACAddress) MarshalJSON() ([]byte, error) {
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

// IsIn checks if a given MAC address belongs to some Ethernet from eths list.
func (m MACAddress) IsIn(eths []*Ethernet) bool {
	mac := m.String()
	for _, eth := range eths {
		if mac == eth.MACAddress.String() {
			return true
		}
	}
	return false
}

// Speed of the Ethernet device (e.g. "10 Mbps").
type Speed string

// Helper lookup table for Speed.MarshalJSON/UnmarshalJSON.
var ethSpeedChoices = map[Speed]int{
	"10 Mbps":       1,
	"100 Mbps":      2,
	"1 Gbps":        3,
	"10 Gbps":       4,
	"40 Gbps":       5,
	"100 Gbps":      6,
	"unknown speed": 11,
}

// MarshalJSON converts Speed from string to an integer code ("choice" in Django
// REST Framework terminology) - for a list of valid values, see the output of
// OPTIONS request on Ralph's API ethernets endpoint.
// This method is necessary, because Ralph's API returns Speed as a string
// (e.g. "10 Mbps"), but accepts it as an int code (choice).
func (s Speed) MarshalJSON() ([]byte, error) {
	speed := ethSpeedChoices[s]
	if speed == 0 {
		return []byte{}, fmt.Errorf("error marshaling Speed: unknown speed: %s", s)
	}
	data, err := json.Marshal(speed)
	if err != nil {
		return []byte{}, fmt.Errorf("error marshaling Speed: %v", err)
	}
	return data, nil
}

// Component is an interface type for Ethernet, Memory etc.
type Component interface {
	IsEqualTo(c Component) bool
	String() string
}

// EthernetList represents the shape of data returned by Ralph for Ethernet endpoint.
type EthernetList struct {
	Count   int
	Results []Ethernet
}

// Ethernet represents a network card on a given host linked to it via
// BaseObject. Speed field is given as int because it should correspond with
// OPTIONS on Ethernet API endpoint (see ethSpeedChoices).
type Ethernet struct {
	ID              int        `json:"id"`
	BaseObject      BaseObject `json:"base_object"`
	MACAddress      MACAddress `json:"mac"`
	ModelName       string     `json:"model_name"`
	Speed           Speed      `json:"speed"`
	FirmwareVersion string     `json:"firmware_version"`
}

func (e Ethernet) String() string {
	return fmt.Sprintf("Ethernet{id: %d, base_object_id: %d, mac: %s, model_name: %s, speed: %s, firmware_version: %s}",
		e.ID, e.BaseObject.ID, e.MACAddress, e.ModelName, e.Speed, e.FirmwareVersion)
}

// IsEqualTo implements Component interface. This method compares two Ethernet
// objects for equality. Please note that Ethernet.ID *is not* taken into
// account here!
func (e Ethernet) IsEqualTo(c Component) bool {
	switch ee := c.(type) {
	case *Ethernet:
		switch {
		case e.BaseObject.ID != ee.BaseObject.ID:
			return false
		case e.MACAddress.String() != ee.MACAddress.String():
			return false
		case e.ModelName != ee.ModelName:
			return false
		case e.Speed != ee.Speed:
			return false
		case e.FirmwareVersion != ee.FirmwareVersion:
			return false
		default:
			return true
		}
	case Ethernet:
		// There's some slightly annoying inconsistency in Go - you can access e
		// fields whenever e is Ethernet or *Ethernet (and that's really
		// convenient), but when it comes to type assertions, Ethernet and
		// *Ethernet are different things - and of course, they are, there's no
		// doubt about that, but...
		// Anyway, I think it's still better to do some "acrobatics" like this
		// recursive call below, than make an exact copy of the body of
		// *Ethernet case...
		// Or maybe automatic code generation would provide a better solution
		// here..?
		// TODO(xor-xor): consider using github.com/clipperhouse/gen for code
		// generation.
		return e.IsEqualTo(&ee)
	default:
		return false
	}
}

// CompareEthernets compares two sets of Ethernet objects (old and new) and
// creates a Diff holding detected changes.
func CompareEthernets(old, new []*Ethernet) (*Diff, error) {
	var all []*Ethernet
	var create, update, delete []*DiffComponent
	all = append(all, old...)
	all = append(all, new...)

	for _, eth := range all {
		switch mac := eth.MACAddress; {
		case mac.IsIn(new) && !mac.IsIn(old):
			d, err := NewDiffComponent(eth)
			if err != nil {
				return nil, err
			}
			create = append(create, d)
		case mac.IsIn(old) && !mac.IsIn(new):
			d, err := NewDiffComponent(eth)
			if err != nil {
				return nil, err
			}
			delete = append(delete, d)
		default: // is present both in new and old
			if eth.ID == 0 {
				// We need to find matching eth in old and assign its ID to the
				// new one to make an update.
				var ethOld *Ethernet
				for _, ethOld = range old {
					if ethOld.MACAddress.String() == eth.MACAddress.String() {
						eth.ID = ethOld.ID
						break
					}
				}
				if eth.ID == 0 {
					// This really shouldn't happen.
					return nil, fmt.Errorf("error while preparing Ethernet with MAC %s for update", eth.MACAddress)
				}
				// Exclude eth equal to ethOld (no need to make an update in
				// such case).
				if !eth.IsEqualTo(ethOld) {
					d, err := NewDiffComponent(eth)
					if err != nil {
						return nil, err
					}
					update = append(update, d)
				}
			}
		}
	}
	return &Diff{
		Create: create,
		Delete: delete,
		Update: update,
	}, nil
}

// MemoryList represents the shape of data returned by Ralph for Memory endpoint.
type MemoryList struct {
	Count   int
	Results []Memory
}

// Memory represents RAM installed on a given host.
type Memory struct {
	ID         int        `json:"id"`
	BaseObject BaseObject `json:"base_object"`
	ModelName  string     `json:"model_name"`
	Size       int        `json:"size"`
	Speed      int        `json:"speed"`
}

func (m Memory) String() string {
	return fmt.Sprintf("Memory{id: %d, base_object_id: %d, model_name: %s, size: %d, speed: %d}",
		m.ID, m.BaseObject.ID, m.ModelName, m.Size, m.Speed)
}

// MarshalJSON serializes Memory into []byte.
func (m Memory) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(map[string]interface{}{
		"id":          m.ID,
		"base_object": m.BaseObject.ID,
		"model_name":  m.ModelName,
		"size":        m.Size,
		"speed":       m.Speed,
	})
	if err != nil {
		return []byte{}, fmt.Errorf("error marshaling Memory: %v", err)
	}
	return data, nil
}

// IsEqualTo implements Component interface. This method compares two Memory
// objects for equality. Please note that Memory.ID *is not* taken into account
// here!
func (m Memory) IsEqualTo(c Component) bool {
	switch mm := c.(type) {
	case *Memory:
		switch {
		case m.BaseObject.ID != mm.BaseObject.ID:
			return false
		case m.ModelName != mm.ModelName:
			return false
		case m.Size != mm.Size:
			return false
		case m.Speed != mm.Speed:
			return false
		default:
			return true
		}
	case Memory:
		return m.IsEqualTo(&mm)
	default:
		return false
	}
}

// CompareMemory compares two sets of Memory objects (old and new) and creates a
// Diff holding detected changes.
func CompareMemory(old, new []*Memory) (*Diff, error) {
	// Since Memory instances are not unique (i.e., w/o considering Memory.ID),
	// we won't be using Diff.Update here.
	var create, delete []*DiffComponent

	// Keys for these maps are constructed from all Memory field values *except* ID.
	counterOld := make(map[string]int)
	counterNew := make(map[string]int)
	oldAsMap := make(map[string][]*Memory)
	newAsMap := make(map[string][]*Memory)

	// Populate oldAsMap/newAsMap.
	for _, m := range old {
		k := fmt.Sprintf("%d__%s__%d__%d",
			m.BaseObject.ID, strings.Replace(m.ModelName, " ", "_", -1), m.Size, m.Speed)
		counterOld[k]++
		oldAsMap[k] = append(oldAsMap[k], m)
	}
	for _, m := range new {
		k := fmt.Sprintf("%d__%s__%d__%d",
			m.BaseObject.ID, strings.Replace(m.ModelName, " ", "_", -1), m.Size, m.Speed)
		counterNew[k]++
		newAsMap[k] = append(newAsMap[k], m)
	}

	if len(new) == 0 {
		for _, m := range old {
			d, err := NewDiffComponent(m)
			if err != nil {
				return nil, err
			}
			delete = append(delete, d)
		}
		return &Diff{
			Create: []*DiffComponent{},
			Delete: delete,
			Update: []*DiffComponent{},
		}, nil
	}

	// Find the differences between counterNew and counterOld and populate
	// Diff.Create and Diff.Delete lists.
	for k, v := range counterNew {
		switch {
		case v > counterOld[k]:
			// Create (v - counterOld[k]) instances of k.
			for i := 0; i < v-counterOld[k]; i++ {
				d, err := NewDiffComponent(newAsMap[k][0])
				if err != nil {
					return nil, err
				}
				create = append(create, d)
			}
		case v < counterOld[k]:
			// Delete (counterOld[k] - v ) instances of k.
			for i := 0; i < counterOld[k]-v; i++ {
				d, err := NewDiffComponent(oldAsMap[k][i])
				if err != nil {
					return nil, err
				}
				delete = append(delete, d)
			}
		}
	}

	// We also need to make sure that keys which are only in counterOld will be
	// deleted.
	for k, v := range counterOld {
		notPresentInNew := false
		for kk := range counterNew {
			if k == kk {
				continue
			}
			notPresentInNew = true
		}
		if notPresentInNew {
			// Delete v instances of k.
			for i := 0; i < v; i++ {
				d, err := NewDiffComponent(oldAsMap[k][i])
				if err != nil {
					return nil, err
				}
				delete = append(delete, d)
			}
		}
	}

	return &Diff{
		Create: create,
		Delete: delete,
		Update: []*DiffComponent{},
	}, nil
}

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

// Disk represents a single hard drive (be it SSD or "normal" one) on a given host.
type Disk struct {
	ModelName    string `json:"model_name"`
	Size         int    `json:"size"`
	SerialNumber string `json:"serial_number"`
}

func (d Disk) String() string {
	return fmt.Sprintf("Disk{model_name: %s, size: %d, sn: %s}",
		d.ModelName, d.Size, d.SerialNumber)
}

// Software represents either firmware or operating system detected on a given
// host.
type Software struct {
	Type    string // possible values: "firmware" and "os"
	Name    string // e.g. "Debian Linux"
	Version string
}

// Model represents a model name of a given physical host (e.g., Dell PowerEdge R620).
type Model struct {
	Name string `json:"model_name"`
}

func (m Model) String() string {
	return fmt.Sprintf("Model{name: %s}", m.Name)
}

// IPAddress is a helper type, i.e. its instances are not meant to be sent to Ralph.
type IPAddress struct {
	Address      string `json:"address"`
	IsMgmt       bool   `json:"is_management"`
	ExposeInDHCP bool   `json:"dhcp_expose"`
	Ethernet     *Ethernet
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
