package main

import (
	"encoding/json"
	"fmt"
)

// Diff represents a set of bulk changes to be made on the same Component
// associated with the same BaseObject (e.g. Ethernet, Memory). These changes
// are grouped in three categories: Create, Update and Delete accounting for
// respective CRUD operations. You can think of Diff as a kind of container for
// DiffComponents.
type Diff struct {
	Create []*DiffComponent
	Update []*DiffComponent
	Delete []*DiffComponent
}

func (d Diff) String() string {
	return fmt.Sprintf("Diff{Create: %s, Update: %s, Delete: %s}",
		d.Create, d.Update, d.Delete)
}

// IsEmpty returns true if a given Diff is empty or false otherwise.
func (d Diff) IsEmpty() bool {
	if len(d.Create) == 0 && len(d.Update) == 0 && len(d.Delete) == 0 {
		return true
	}
	return false
}

// DiffComponent represents a single piece of changes that will be send to
// Ralph. It's main part is the Data field, holding JSON-ised payload that will
// be send to Ralph. Other fields provide convenience shourtcuts for some
// functions/methods (e.g. Component field holds a reference to the original
// object, which frees us from unmarshaling contents of Data field).
type DiffComponent struct {
	ID        int       // ID of the object held in Component field
	Name      string    // name of the component (e.g. Ethernet, Memory)
	Data      []byte    // JSON-ed Component field
	Component Component // reference to the original object
}

// NewDiffComponent creates a DiffComponent based on a given component. Since
// component is an interface type, it can hold both object or pointer, but
// NewDiffComponent handles both of these cases.
func NewDiffComponent(component Component) (*DiffComponent, error) {
	var id int
	var name string
	var data []byte
	var err error
	switch v := component.(type) {
	case *Ethernet:
		id = v.ID
		name = "Ethernet"
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	case Ethernet:
		return NewDiffComponent(&v)
	case *Memory:
		id = v.ID
		name = "Memory"
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	case Memory:
		return NewDiffComponent(&v)
	case *FibreChannelCard:
		id = v.ID
		name = "FibreChannelCard"
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	case FibreChannelCard:
		return NewDiffComponent(&v)
	case *Processor:
		id = v.ID
		name = "Processor"
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	case Processor:
		return NewDiffComponent(&v)
	case *Disk:
		id = v.ID
		name = "Disk"
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	case Disk:
		return NewDiffComponent(&v)
	case *DataCenterAsset:
		id = *v.ID
		name = "DataCenterAsset"
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	case DataCenterAsset:
		return NewDiffComponent(&v)
	default:
		return nil, fmt.Errorf("unknown component: %+v", v)
	}
	return &DiffComponent{
		ID:        id,
		Name:      name,
		Data:      data,
		Component: component,
	}, nil
}

func (d DiffComponent) String() string {
	return fmt.Sprintf("DiffComponent{ID: %d, Name: %s, Data: %s, Component: %s}",
		d.ID, d.Name, string(d.Data), d.Component)
}

// SendDiffToRalph sends a given Diff to Ralph. If dryRun is set to true, then
// no changes will be sent to Ralph. If noOutput is set to true, then all the
// output from this function will be silenced (this is mostly for tests).
// Returned statusCodes slice is meant only to facilitate tests, so don't be
// surprised if you see it ignored somewhere in the source code.
func SendDiffToRalph(client *Client, diff *Diff, dryRun bool, noOutput bool) (statusCodes []int, err error) {

	var send = func(d *DiffComponent, method, endpoint, msg string) (int, error) {
		var code int
		var data []byte
		switch {
		case method == "DELETE":
			data = nil
		default:
			data = d.Data
		}
		if !dryRun {
			code, err = client.SendToRalph(method, endpoint, data)
		}
		if err != nil {
			return code, err
		}
		if !noOutput {
			fmt.Printf("%s %s successfully.\n", d.Component, msg) // TODO(xor-xor): Use logger instead.
		}
		return code, nil
	}

	for _, d := range diff.Create {
		endpoint := APIEndpoints[d.Name]
		code, err := send(d, "POST", endpoint, "created")
		if err != nil {
			return statusCodes, err
		}
		if code != 0 {
			statusCodes = append(statusCodes, code)
		}
	}
	for _, d := range diff.Update {
		endpoint := fmt.Sprintf("%s/%d", APIEndpoints[d.Name], d.ID)
		code, err := send(d, "PATCH", endpoint, "updated")
		if err != nil {
			return statusCodes, err
		}
		if code != 0 {
			statusCodes = append(statusCodes, code)
		}
	}
	for _, d := range diff.Delete {
		endpoint := fmt.Sprintf("%s/%d", APIEndpoints[d.Name], d.ID)
		code, err := send(d, "DELETE", endpoint, "deleted")
		if err != nil {
			return statusCodes, err
		}
		if code != 0 {
			statusCodes = append(statusCodes, code)
		}
	}
	return statusCodes, nil
}
