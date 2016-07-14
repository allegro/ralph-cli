package main

import (
	"strings"
	"testing"

	"github.com/juju/testing/checkers"
)

func TestDiffIsEmpty(t *testing.T) {
	var cases = map[string]struct {
		diff *Diff
		want bool
	}{
		"#0 Is empty": {
			&Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
			true,
		},
		"#1 Is not empty": {
			&Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "Ethernet",
						Data:      []byte(`{"ID":0,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"","speed":11,"firmware_version":""}`),
						Component: &Ethernet{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "", "unknown speed", ""},
					},
				},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
			false,
		},
	}
	for tn, tc := range cases {
		got := tc.diff.IsEmpty()
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}

// TODO(xor-xor): Refactor SendDiffToRalph (or better yet, MockServerClient) to support
// testing scenarios, where multiple, different HTTP status codes are returned.
func TestSendDiffToRalph(t *testing.T) {
	var cases = map[string]struct {
		diff       *Diff
		dryRun     bool
		statusCode int
		want       []int
	}{
		"#0 Empty diff": {
			&Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
			false,
			0, // In this case, statusCode doesn't really matter.
			[]int{},
		},
		"#1 Dry run": {
			&Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:        0,
						Name:      "Ethernet",
						Data:      []byte(`{"ID":0,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"","speed":11,"firmware_version":""}`),
						Component: &Ethernet{0, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "unknown speed", ""},
					},
				},
				Update: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "Ethernet",
						Data:      []byte(`{"ID":1,"base_object":2,"mac":"a1:b2:c3:d4:e5:f6","model_name":"","speed":11,"firmware_version":""}`),
						Component: &Ethernet{1, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "unknown speed", ""},
					},
				},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        2,
						Name:      "Ethernet",
						Data:      []byte(`{"ID":2,"base_object":3,"mac":"aa:aa:aa:aa:aa:aa","model_name":"","speed":11,"firmware_version":""}`),
						Component: &Ethernet{1, BaseObject{3}, macs["aa:aa:aa:aa:aa:aa"], "", "unknown speed", ""},
					},
				},
			},
			true,
			0, // In this case, statusCode doesn't really matter.
			[]int{},
		},
		"#2 Create": {
			&Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:        0,
						Name:      "Ethernet",
						Data:      []byte(`{"ID":0,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"","speed":11,"firmware_version":""}`),
						Component: &Ethernet{0, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "unknown speed", ""},
					},
				},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
			false,
			201,
			[]int{201},
		},
		"#3 Update": {
			&Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "Ethernet",
						Data:      []byte(`{"ID":1,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"","speed":11,"firmware_version":""}`),
						Component: &Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "unknown speed", ""},
					},
				},
				Delete: []*DiffComponent{},
			},
			false,
			200,
			[]int{200},
		},
		"#4 Delete": {
			&Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "Ethernet",
						Data:      []byte(`{"ID":1,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"","speed":11,"firmware_version":""}`),
						Component: &Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "unknown speed", ""},
					},
				},
			},
			false,
			204,
			[]int{204},
		},
	}

	for tn, tc := range cases {
		server, client := MockServerClient(tc.statusCode, `{}`)
		defer server.Close()

		got, err := SendDiffToRalph(client, tc.diff, tc.dryRun, true)
		if err != nil {
			t.Fatalf("%s\nerr: %s", tn, err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestDiffToString(t *testing.T) {
	diff := Diff{
		Create: []*DiffComponent{
			&DiffComponent{
				ID:        1,
				Name:      "Ethernet",
				Data:      []byte(`{"ID":1,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"","speed":11,"firmware_version":""}`),
				Component: &Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "unknown speed", ""},
			},
		},
		Update: []*DiffComponent{},
		Delete: []*DiffComponent{},
	}
	want := `Diff{Create: [DiffComponent{ID: 1, Name: Ethernet, Data: {"ID":1,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"","speed":11,"firmware_version":""}, Component: Ethernet{id: 1, base_object_id: 1, mac: aa:bb:cc:dd:ee:ff, model_name: , speed: unknown speed, firmware_version: }}], Update: [], Delete: []}`

	got := diff.String()
	if got != want {
		t.Errorf("\n got: %v\nwant: %v", got, want)
	}

}

func TestNewDiffComponent(t *testing.T) {
	ethernet := Ethernet{
		ID:              1,
		BaseObject:      BaseObject{1},
		MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
		ModelName:       "Intel(R) Ethernet 10G 4P X520/I350 rNDC",
		Speed:           "10 Gbps",
		FirmwareVersion: "1.1.1",
	}
	memory := Memory{
		ID:         1,
		BaseObject: BaseObject{1},
		ModelName:  "Samsung DDR3 DIMM",
		Size:       16384,
		Speed:      1600,
	}
	fcc := FibreChannelCard{
		ID:              1,
		BaseObject:      BaseObject{1},
		ModelName:       "Saturn-X: LightPulse Fibre Channel Host Adapter",
		Speed:           "4 Gbit",
		WWN:             "aabbccddeeff0011",
		FirmwareVersion: "1.1.1",
	}
	proc := Processor{
		ID:         1,
		BaseObject: BaseObject{1},
		ModelName:  "Intel(R) Xeon(R)",
		Speed:      2600,
		Cores:      8,
	}
	disk := Disk{
		ID:              1,
		BaseObject:      BaseObject{1},
		ModelName:       "ATA Samsung SSD 840",
		Size:            476,
		SerialNumber:    "S1234",
		Slot:            1,
		FirmwareVersion: "1.1.1",
	}
	var cases = map[string]struct {
		component Component
		want      *DiffComponent
		errMsg    string
	}{
		"#0 Unknown component": {
			component: FakeComponent{},
			want:      nil,
			errMsg:    "unknown component:",
		},
		"#1 Ethernet": {
			component: ethernet,
			want: &DiffComponent{
				ID:        1,
				Name:      "Ethernet",
				Data:      []byte(`{"id":1,"base_object":1,"mac":"a1:b2:c3:d4:e5:f6","model_name":"Intel(R) Ethernet 10G 4P X520/I350 rNDC","speed":4,"firmware_version":"1.1.1"}`),
				Component: &ethernet,
			},
			errMsg: "",
		},
		"#2 Ethernet as a pointer": {
			component: &ethernet,
			want: &DiffComponent{
				ID:        1,
				Name:      "Ethernet",
				Data:      []byte(`{"id":1,"base_object":1,"mac":"a1:b2:c3:d4:e5:f6","model_name":"Intel(R) Ethernet 10G 4P X520/I350 rNDC","speed":4,"firmware_version":"1.1.1"}`),
				Component: &ethernet,
			},
			errMsg: "",
		},
		"#3 Memory": {
			component: memory,
			want: &DiffComponent{
				ID:        1,
				Name:      "Memory",
				Data:      []byte(`{"base_object":1,"id":1,"model_name":"Samsung DDR3 DIMM","size":16384,"speed":1600}`),
				Component: &memory,
			},
			errMsg: "",
		},
		"#4 Memory as a pointer": {
			component: &memory,
			want: &DiffComponent{
				ID:        1,
				Name:      "Memory",
				Data:      []byte(`{"base_object":1,"id":1,"model_name":"Samsung DDR3 DIMM","size":16384,"speed":1600}`),
				Component: &memory,
			},
			errMsg: "",
		},
		"#5 FibreChannelCard": {
			component: fcc,
			want: &DiffComponent{
				ID:        1,
				Name:      "FibreChannelCard",
				Data:      []byte(`{"id":1,"base_object":1,"model_name":"Saturn-X: LightPulse Fibre Channel Host Adapter","speed":3,"wwn":"aabbccddeeff0011","firmware_version":"1.1.1"}`),
				Component: &fcc,
			},
			errMsg: "",
		},
		"#6 FibreChannelCard as a pointer": {
			component: &fcc,
			want: &DiffComponent{
				ID:        1,
				Name:      "FibreChannelCard",
				Data:      []byte(`{"id":1,"base_object":1,"model_name":"Saturn-X: LightPulse Fibre Channel Host Adapter","speed":3,"wwn":"aabbccddeeff0011","firmware_version":"1.1.1"}`),
				Component: &fcc,
			},
			errMsg: "",
		},
		"#7 Processor": {
			component: proc,
			want: &DiffComponent{
				ID:        1,
				Name:      "Processor",
				Data:      []byte(`{"id":1,"base_object":1,"model_name":"Intel(R) Xeon(R)","speed":2600,"cores":8}`),
				Component: &proc,
			},
			errMsg: "",
		},
		"#8 Processor as a pointer": {
			component: &proc,
			want: &DiffComponent{
				ID:        1,
				Name:      "Processor",
				Data:      []byte(`{"id":1,"base_object":1,"model_name":"Intel(R) Xeon(R)","speed":2600,"cores":8}`),
				Component: &proc,
			},
			errMsg: "",
		},
		"#9 Disk": {
			component: disk,
			want: &DiffComponent{
				ID:        1,
				Name:      "Disk",
				Data:      []byte(`{"id":1,"base_object":1,"model_name":"ATA Samsung SSD 840","size":476,"serial_number":"S1234","slot":1,"firmware_version":"1.1.1"}`),
				Component: &disk,
			},
			errMsg: "",
		},
		"#10 Disk as a pointer": {
			component: &disk,
			want: &DiffComponent{
				ID:        1,
				Name:      "Disk",
				Data:      []byte(`{"id":1,"base_object":1,"model_name":"ATA Samsung SSD 840","size":476,"serial_number":"S1234","slot":1,"firmware_version":"1.1.1"}`),
				Component: &disk,
			},
			errMsg: "",
		},
	}
	for tn, tc := range cases {
		got, err := NewDiffComponent(tc.component)
		switch {
		case tc.errMsg != "":
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("%s\ndidn't get expected string: %q in err msg: %q", tn, tc.errMsg, err)
			}
		default:
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if eq, err := checkers.DeepEqual(got, tc.want); !eq {
				t.Errorf("%s\n%s", tn, err)
			}
		}
	}
}
