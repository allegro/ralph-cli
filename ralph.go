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
		return nil, fmt.Errorf("error unmarshaling base object: %v", err)
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
		return nil, fmt.Errorf("error unmarshaling Ethernet: %v", err)
	}
	ethsPtrs := make([]*Ethernet, eths.Count)
	for i := 0; i < eths.Count; i++ {
		ethsPtrs[i] = &eths.Results[i]
	}
	return ethsPtrs, nil
}

// GetMemory fetches Memory objects associated with given BaseObject.
func (b BaseObject) GetMemory(c *Client) ([]*Memory, error) {
	// Hardcoding such things as the limit below is generally a bad idea, but in
	// this particular case it won't hurt (the maximum number of results that we
	// expect from 'memory' endpoint is like 16 or maaaaaybe 32; and BTW, the
	// default limit in Ralph is just 10).
	q := fmt.Sprintf("base_object=%d&limit=100", b.ID)
	rawBody, err := c.GetFromRalph(APIEndpoints["Memory"], q)
	if err != nil {
		return nil, err
	}
	var mems MemoryList
	if err := json.Unmarshal(rawBody, &mems); err != nil {
		return nil, fmt.Errorf("error unmarshaling Memory: %v", err)
	}
	memsPtrs := make([]*Memory, mems.Count)
	for i := 0; i < mems.Count; i++ {
		memsPtrs[i] = &mems.Results[i]
	}
	return memsPtrs, nil
}

// GetFibreChannelCards fetches FibreChannelCard objects associated with given
// BaseObject.
func (b BaseObject) GetFibreChannelCards(c *Client) ([]*FibreChannelCard, error) {
	q := fmt.Sprintf("base_object=%d", b.ID)
	rawBody, err := c.GetFromRalph(APIEndpoints["FibreChannelCard"], q)
	if err != nil {
		return nil, err
	}
	var cards FibreChannelCardList
	if err := json.Unmarshal(rawBody, &cards); err != nil {
		return nil, fmt.Errorf("error unmarshaling FibreChannelCard: %v", err)
	}
	cardsPtrs := make([]*FibreChannelCard, cards.Count)
	for i := 0; i < cards.Count; i++ {
		cardsPtrs[i] = &cards.Results[i]
	}
	return cardsPtrs, nil
}

// GetProcessors fetches Processor objects associated with given
// BaseObject.
func (b BaseObject) GetProcessors(c *Client) ([]*Processor, error) {
	q := fmt.Sprintf("base_object=%d", b.ID)
	rawBody, err := c.GetFromRalph(APIEndpoints["Processor"], q)
	if err != nil {
		return nil, err
	}
	var procs ProcessorList
	if err := json.Unmarshal(rawBody, &procs); err != nil {
		return nil, fmt.Errorf("error unmarshaling Processor: %v", err)
	}
	procsPtrs := make([]*Processor, procs.Count)
	for i := 0; i < procs.Count; i++ {
		procsPtrs[i] = &procs.Results[i]
	}
	return procsPtrs, nil
}

// GetDisks fetches Disk objects associated with given BaseObject.
func (b BaseObject) GetDisks(c *Client) ([]*Disk, error) {
	q := fmt.Sprintf("base_object=%d", b.ID)
	rawBody, err := c.GetFromRalph(APIEndpoints["Disk"], q)
	if err != nil {
		return nil, err
	}
	var disks DiskList
	if err := json.Unmarshal(rawBody, &disks); err != nil {
		return nil, fmt.Errorf("error unmarshaling Disk: %v", err)
	}
	disksPtrs := make([]*Disk, disks.Count)
	for i := 0; i < disks.Count; i++ {
		disksPtrs[i] = &disks.Results[i]
	}
	return disksPtrs, nil
}

// GetDataCenterAsset fetches DataCenterAsset object associated with given
// BaseObject. Please note, that there will be only one such object, hence we do
// not return an array here.
func (b BaseObject) GetDataCenterAsset(c *Client) (*DataCenterAsset, error) {
	endpoint := fmt.Sprintf("%s/%d", APIEndpoints["DataCenterAsset"], b.ID)
	rawBody, err := c.GetFromRalph(endpoint, "")
	if err != nil {
		return nil, err
	}
	var dcAsset DataCenterAsset
	if err := json.Unmarshal(rawBody, &dcAsset); err != nil {
		return nil, fmt.Errorf("error unmarshaling DataCenterAsset: %v", err)
	}
	return &dcAsset, nil
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

// EthSpeed is the speed of the Ethernet device (e.g. "10 Mbps").
type EthSpeed string

// Helper lookup table for EthSpeed.MarshalJSON/UnmarshalJSON.
var ethSpeedChoices = map[EthSpeed]int{
	"10 Mbps":       1,
	"100 Mbps":      2,
	"1 Gbps":        3,
	"10 Gbps":       4,
	"40 Gbps":       5,
	"100 Gbps":      6,
	"unknown speed": 11,
}

// MarshalJSON converts EthSpeed from string to an integer code ("choice" in
// Django REST Framework terminology) - for a list of valid values, see the
// output of OPTIONS request on Ralph's API ethernets endpoint.
// This method is necessary, because Ralph's API returns this speed as a string
// (e.g. "10 Mbps"), but accepts it as an int code (choice).
func (s EthSpeed) MarshalJSON() ([]byte, error) {
	speed := ethSpeedChoices[s]
	if speed == 0 {
		return []byte{}, fmt.Errorf("error marshaling EthSpeed: unknown speed: %s", s)
	}
	data, err := json.Marshal(speed)
	if err != nil {
		return []byte{}, fmt.Errorf("error marshaling EthSpeed: %v", err)
	}
	return data, nil
}

// FCCSpeed is the speed of the FibreChannelCard (e.g. "32 Gbit").
type FCCSpeed string

// Helper lookup table for FCCSpeed.MarshalJSON/UnmarshalJSON.
var fccSpeedChoices = map[FCCSpeed]int{
	"1 Gbit":        1,
	"2 Gbit":        2,
	"4 Gbit":        3,
	"8 Gbit":        4,
	"16 Gbit":       5,
	"32 Gbit":       6,
	"unknown speed": 11,
}

// MarshalJSON converts FCCSpeed from string to an integer code ("choice" in
// Django REST Framework terminology) - for a list of valid values, see the
// output of OPTIONS request on Ralph's API ethernets endpoint.
// This method is necessary, because Ralph's API returns this speed as a string
// (e.g. "32 Gbit"), but accepts it as an int code (choice).
func (s FCCSpeed) MarshalJSON() ([]byte, error) {
	speed := fccSpeedChoices[s]
	if speed == 0 {
		return []byte{}, fmt.Errorf("error marshaling FCCSpeed: unknown speed: %s", s)
	}
	data, err := json.Marshal(speed)
	if err != nil {
		return []byte{}, fmt.Errorf("error marshaling FCCSpeed: %v", err)
	}
	return data, nil
}

// DiskSlotNumber is a helper type for Disk.Slot field. Apart from "normal" slot
// numbers (e.g. 0, 1, 2...) it may have a special value -1, which should be
// interpreted as nil.  The only purpose of this type is to circuvement possible
// collisions between 0 as a "normal" slot number vs. 0 as a zero value in Go.
// See also DataCenterAsset datatype for alternative approach (struct fields as
// pointers facilitating PATCH-ing a resource).
type DiskSlotNumber int

// MarshalJSON serializes DiskSlotNumber to []byte.
func (d DiskSlotNumber) MarshalJSON() ([]byte, error) {
	switch {
	case d == -1:
		return []byte("null"), nil
	default:
		return json.Marshal(int(d))
	}
}

// UnmarshalJSON deserializes DiskSlotNumber from []byte.
func (d *DiskSlotNumber) UnmarshalJSON(data []byte) error {
	switch {
	case string(data) == "null":
		*d = -1
	default:
		var slot int
		if err := json.Unmarshal(data, &slot); err != nil {
			return fmt.Errorf("error unmarshaling DiskSlotNumber: %s", err)
		}
		*d = (DiskSlotNumber)(slot)
	}
	return nil
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
// BaseObject.
type Ethernet struct {
	ID              int        `json:"id"`
	BaseObject      BaseObject `json:"base_object"`
	MACAddress      MACAddress `json:"mac"`
	ModelName       string     `json:"model_name"`
	Speed           EthSpeed   `json:"speed"`
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
		case v == counterOld[k]:
			continue
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
		presentInNew := false
		for kk := range counterNew {
			if k == kk {
				presentInNew = true
				break
			}
		}
		if !presentInNew {
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

// FibreChannelCardList represents the shape of data returned by Ralph for
// FibreChannelCard endpoint.
type FibreChannelCardList struct {
	Count   int
	Results []FibreChannelCard
}

// FibreChannelCard represents a single fibre channel card/controller on a given
// host.
type FibreChannelCard struct {
	ID              int        `json:"id"`
	BaseObject      BaseObject `json:"base_object"`
	ModelName       string     `json:"model_name"`
	Speed           FCCSpeed   `json:"speed"`
	WWN             string     `json:"wwn"`
	FirmwareVersion string     `json:"firmware_version"`
}

func (f FibreChannelCard) String() string {
	return fmt.Sprintf("FibreChannelCard{id: %d, base_object_id: %d, model_name: %s, speed: %s, wwn: %s, firmware_version: %s}",
		f.ID, f.BaseObject.ID, f.ModelName, f.Speed, f.WWN, f.FirmwareVersion)
}

// IsEqualTo implements Component interface. This method compares two
// FibreChannelCard objects for equality. Their IDs are not taken into account.
func (f FibreChannelCard) IsEqualTo(c Component) bool {
	switch ff := c.(type) {
	case *FibreChannelCard:
		switch {
		case f.BaseObject.ID != ff.BaseObject.ID:
			return false
		case f.ModelName != ff.ModelName:
			return false
		case f.Speed != ff.Speed:
			return false
		case f.WWN != ff.WWN:
			return false
		case f.FirmwareVersion != ff.FirmwareVersion:
			return false
		default:
			return true
		}
	case FibreChannelCard:
		return f.IsEqualTo(&ff)
	default:
		return false
	}
}

// CompareFibreChannelCards compares two sets of FibreChannelCard objects (old
// and new) and creates a Diff holding detected changes.
func CompareFibreChannelCards(old, new []*FibreChannelCard) (*Diff, error) {
	var create, delete []*DiffComponent

	// At first, it may seem that FibreChannelCards can be compared as unique
	// objects, thanks to WWN, but unfortunately, this field will be empty quite
	// often, so we have to compare them in the same way as with Memory.

	// Keys for these maps are constructed from all FibreChannelCard field
	// values *except* ID.
	counterOld := make(map[string]int)
	counterNew := make(map[string]int)
	oldAsMap := make(map[string][]*FibreChannelCard)
	newAsMap := make(map[string][]*FibreChannelCard)

	// Populate oldAsMap/newAsMap.
	for _, fc := range old {
		k := fmt.Sprintf("%d__%s__%s__%s__%s",
			fc.BaseObject.ID, strings.Replace(fc.ModelName, " ", "_", -1), fc.Speed, fc.WWN, fc.FirmwareVersion)
		counterOld[k]++
		oldAsMap[k] = append(oldAsMap[k], fc)
	}
	for _, fc := range new {
		k := fmt.Sprintf("%d__%s__%s__%s__%s",
			fc.BaseObject.ID, strings.Replace(fc.ModelName, " ", "_", -1), fc.Speed, fc.WWN, fc.FirmwareVersion)
		counterNew[k]++
		newAsMap[k] = append(newAsMap[k], fc)
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
		case v == counterOld[k]:
			continue
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
		presentInNew := false
		for kk := range counterNew {
			if k == kk {
				presentInNew = true
				break
			}
		}
		if !presentInNew {
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

// ProcessorList represents the shape of data returned by Ralph for Processor
// endpoint.
type ProcessorList struct {
	Count   int
	Results []Processor
}

// Processor represents a single processor on a given host.
type Processor struct {
	ID         int        `json:"id"`
	BaseObject BaseObject `json:"base_object"`
	ModelName  string     `json:"model_name"`
	Speed      int        `json:"speed"`
	Cores      int        `json:"cores"`
}

func (p Processor) String() string {
	return fmt.Sprintf("Processor{id: %d, base_object_id: %d, model_name: %s, speed: %d, cores: %d}",
		p.ID, p.BaseObject.ID, p.ModelName, p.Speed, p.Cores)
}

// IsEqualTo implements Component interface. This method compares two Processor
// objects for equality. Please note that Processor.ID *is not* taken into
// account here!
func (p Processor) IsEqualTo(c Component) bool {
	switch pp := c.(type) {
	case *Processor:
		switch {
		case p.BaseObject.ID != pp.BaseObject.ID:
			return false
		case p.ModelName != pp.ModelName:
			return false
		case p.Speed != pp.Speed:
			return false
		case p.Cores != pp.Cores:
			return false
		default:
			return true
		}
	case Processor:
		return p.IsEqualTo(&pp)
	default:
		return false
	}
}

// CompareProcessors compares two sets of Processor objects (old and new) and
// creates a Diff holding detected changes.
func CompareProcessors(old, new []*Processor) (*Diff, error) {
	// Since Processor instances are not unique (i.e., w/o considering Processor.ID),
	// we won't be using Diff.Update here.
	var create, delete []*DiffComponent

	// Keys for these maps are constructed from all Processor field values *except* ID.
	counterOld := make(map[string]int)
	counterNew := make(map[string]int)
	oldAsMap := make(map[string][]*Processor)
	newAsMap := make(map[string][]*Processor)

	// Populate oldAsMap/newAsMap.
	for _, p := range old {
		k := fmt.Sprintf("%d__%s__%d__%d",
			p.BaseObject.ID, strings.Replace(p.ModelName, " ", "_", -1), p.Speed, p.Cores)
		counterOld[k]++
		oldAsMap[k] = append(oldAsMap[k], p)
	}
	for _, p := range new {
		k := fmt.Sprintf("%d__%s__%d__%d",
			p.BaseObject.ID, strings.Replace(p.ModelName, " ", "_", -1), p.Speed, p.Cores)
		counterNew[k]++
		newAsMap[k] = append(newAsMap[k], p)
	}

	if len(new) == 0 {
		for _, p := range old {
			d, err := NewDiffComponent(p)
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
		case v == counterOld[k]:
			continue
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
		presentInNew := false
		for kk := range counterNew {
			if k == kk {
				presentInNew = true
				break
			}
		}
		if !presentInNew {
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

// DiskList represents the shape of data returned by Ralph for Disk endpoint.
type DiskList struct {
	Count   int
	Results []Disk
}

// Disk represents a single disk drive (be it HDD or SSD) on a given host.
type Disk struct {
	ID              int            `json:"id"`
	BaseObject      BaseObject     `json:"base_object"`
	ModelName       string         `json:"model_name"`
	Size            int            `json:"size"`
	SerialNumber    string         `json:"serial_number"`
	Slot            DiskSlotNumber `json:"slot"`
	FirmwareVersion string         `json:"firmware_version"`
}

func (d Disk) String() string {
	var slot string
	if d.Slot != -1 {
		slot = fmt.Sprintf("%d", d.Slot)
	}
	return fmt.Sprintf("Disk{id: %d, base_object_id: %d, model_name: %s, size: %d, serial_number: %s, slot: %s, firmware_version: %s}",
		d.ID, d.BaseObject.ID, d.ModelName, d.Size, d.SerialNumber, slot, d.FirmwareVersion)
}

// IsEqualTo implements Component interface. This method compares two Disk
// objects for equality. Please note that Disk.ID *is not* taken into account
// here!
func (d Disk) IsEqualTo(c Component) bool {
	switch dd := c.(type) {
	case *Disk:
		switch {
		case d.BaseObject.ID != dd.BaseObject.ID:
			return false
		case d.ModelName != dd.ModelName:
			return false
		case d.Size != dd.Size:
			return false
		case d.SerialNumber != dd.SerialNumber:
			return false
		case d.Slot != dd.Slot:
			return false
		case d.FirmwareVersion != dd.FirmwareVersion:
			return false
		default:
			return true
		}
	case Disk:
		return d.IsEqualTo(&dd)
	default:
		return false
	}
}

// CompareDisks compares two sets of Disk objects (old and new) and creates a
// Diff holding detected changes.
func CompareDisks(old, new []*Disk) (*Diff, error) {
	// Since Disk instances are not unique (i.e., w/o considering Disk.ID),
	// we won't be using Diff.Update here.
	var create, delete []*DiffComponent

	// Keys for these maps are constructed from all Disk field values *except* ID.
	counterOld := make(map[string]int)
	counterNew := make(map[string]int)
	oldAsMap := make(map[string][]*Disk)
	newAsMap := make(map[string][]*Disk)

	// Populate oldAsMap/newAsMap.
	for _, d := range old {
		k := fmt.Sprintf("%d__%s__%d__%s__%d__%s",
			d.BaseObject.ID, strings.Replace(d.ModelName, " ", "_", -1), d.Size, d.SerialNumber, d.Slot, d.FirmwareVersion)
		counterOld[k]++
		oldAsMap[k] = append(oldAsMap[k], d)
	}
	for _, d := range new {
		k := fmt.Sprintf("%d__%s__%d__%s__%d__%s",
			d.BaseObject.ID, strings.Replace(d.ModelName, " ", "_", -1), d.Size, d.SerialNumber, d.Slot, d.FirmwareVersion)
		counterNew[k]++
		newAsMap[k] = append(newAsMap[k], d)
	}

	if len(new) == 0 {
		for _, d := range old {
			dc, err := NewDiffComponent(d)
			if err != nil {
				return nil, err
			}
			delete = append(delete, dc)
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
		case v == counterOld[k]:
			continue
		case v > counterOld[k]:
			// Create (v - counterOld[k]) instances of k.
			for i := 0; i < v-counterOld[k]; i++ {
				dc, err := NewDiffComponent(newAsMap[k][0])
				if err != nil {
					return nil, err
				}
				create = append(create, dc)
			}
		case v < counterOld[k]:
			// Delete (counterOld[k] - v ) instances of k.
			for i := 0; i < counterOld[k]-v; i++ {
				dc, err := NewDiffComponent(oldAsMap[k][i])
				if err != nil {
					return nil, err
				}
				delete = append(delete, dc)
			}
		}
	}

	// We also need to make sure that keys which are only in counterOld will be
	// deleted.
	for k, v := range counterOld {
		presentInNew := false
		for kk := range counterNew {
			if k == kk {
				presentInNew = true
				break
			}
		}
		if !presentInNew {
			// Delete v instances of k.
			for i := 0; i < v; i++ {
				dc, err := NewDiffComponent(oldAsMap[k][i])
				if err != nil {
					return nil, err
				}
				delete = append(delete, dc)
			}
		}
	}

	return &Diff{
		Create: create,
		Delete: delete,
		Update: []*DiffComponent{},
	}, nil
}

// DataCenterAsset is meant only for updating firmware_version and bios_version
// fields on Ralph's DataCenterAsset model, putting ScanResult.Model into
// Remarks and for determining correctness of SerialNumber (detected vs. stored
// in Ralph). This datatype should be sent to Ralph only with PATCH method. It
// is also an experiment with the approach presented in this article:
// https://willnorris.com/2014/05/go-rest-apis-and-pointers (struct fields as
// pointers facilitating PATCH-ing a resource).
type DataCenterAsset struct {
	ID              *int    `json:"id,omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty"`
	BIOSVersion     *string `json:"bios_version,omitempty"`
	Remarks         *string `json:"remarks,omitempty"`
	SerialNumber    *string `json:"sn,omitempty"`
}

// String for DataCenterAsset will present only the fields that are not nil.
func (a DataCenterAsset) String() string {
	var str string
	if a.ID != nil {
		str += fmt.Sprintf("id: %d, ", *a.ID)
	}
	if a.FirmwareVersion != nil {
		str += fmt.Sprintf("firmware_version: %s, ", *a.FirmwareVersion)
	}
	if a.BIOSVersion != nil {
		str += fmt.Sprintf("bios_version: %s, ", *a.BIOSVersion)
	}
	if a.Remarks != nil {
		remarks := strings.Replace(*a.Remarks, "\r\n", " ", -1)
		remarks = strings.Replace(remarks, "\n", " ", -1)
		str += fmt.Sprintf("remarks: %s, ", remarks)
	}
	if a.SerialNumber != nil {
		str += fmt.Sprintf("sn: %s, ", *a.SerialNumber)
	}
	return fmt.Sprintf("DataCenterAsset{%s}", strings.TrimSuffix(str, ", "))
}

// IsEqualTo implements Component interface. This method compares two
// DataCenterAsset objects for equality. Please note that DataCenterAsset.ID *is
// not* taken into account here, and that this method's body slightly differs
// from other components because all DataCenterAsset fields are pointers.
func (a DataCenterAsset) IsEqualTo(c Component) bool {
	switch aa := c.(type) {
	case *DataCenterAsset:
		switch {
		// Checking if exactly one of the pointers is nil.
		case a.FirmwareVersion == nil && aa.FirmwareVersion != nil,
			a.FirmwareVersion != nil && aa.FirmwareVersion == nil:
			return false
		case a.BIOSVersion == nil && aa.BIOSVersion != nil,
			a.BIOSVersion != nil && aa.BIOSVersion == nil:
			return false
		case a.Remarks == nil && aa.Remarks != nil,
			a.Remarks != nil && aa.Remarks == nil:
			return false
		case a.SerialNumber == nil && aa.SerialNumber != nil,
			a.SerialNumber != nil && aa.SerialNumber == nil:
			return false

		// Checking if both pointers are not nil and the contents are different.
		case a.FirmwareVersion != nil && aa.FirmwareVersion != nil &&
			*a.FirmwareVersion != *aa.FirmwareVersion:
			return false
		case a.BIOSVersion != nil && aa.BIOSVersion != nil &&
			*a.BIOSVersion != *aa.BIOSVersion:
			return false
		case a.Remarks != nil && aa.Remarks != nil &&
			*a.Remarks != *aa.Remarks:
			return false
		case a.SerialNumber != nil && aa.SerialNumber != nil &&
			*a.SerialNumber != *aa.SerialNumber:
			return false

		default:
			return true
		}
	case DataCenterAsset:
		return a.IsEqualTo(&aa)
	default:
		return false
	}
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
		return nil, fmt.Errorf("error unmarshaling IPAddress: %v", err)
	}
	return &addrs, nil
}
