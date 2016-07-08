package main

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/juju/testing/checkers"
)

var ralphTestFixturesDir = "./ralph_test_fixtures"

func TestNewAddr(t *testing.T) {
	var cases = []struct {
		input string
		want  Addr
	}{
		{"10.20.30.40", Addr("10.20.30.40")},
		{"10.20.30.40.50", ""},
		{"10.20.30.257", ""},
		{"255.255.255.255", Addr("255.255.255.255")},
		{"0.0.0.0", Addr("0.0.0.0")},
		{"allegro.pl", Addr("allegro.pl")},
		{"certainly.does.not.exist", ""},
		{"", ""},
	}
	for tn, tc := range cases {
		got, _ := NewAddr(tc.input)
		if got != tc.want {
			t.Errorf("#%d\n got: %q\nwant: %q", tn, got, tc.want)
		}
	}
}

func TestEthernetIsEqualTo(t *testing.T) {
	var cases = map[string]struct {
		eth  *Ethernet
		comp Component
		want bool
	}{
		"#0 All equal": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			true,
		},
		"#1 All different": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&Ethernet{2, BaseObject{2}, macs["aa:aa:aa:aa:aa:aa"], "Intel Corporation 82599EB 10-Gigabit SFI/SFP", "10 Gbps", "1.1.1"},
			false,
		},
		"#2 Different BaseObject.ID": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&Ethernet{1, BaseObject{2}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			false,
		},
		"#3 Different MACAddress": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&Ethernet{1, BaseObject{1}, macs["aa:aa:aa:aa:aa:aa"], "", "", ""},
			false,
		},
		"#4 Different Model": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "Intel Corporation 82599EB 10-Gigabit SFI/SFP", "", ""},
			false,
		},
		"#5 Different Speed": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "10 Gbps", ""},
			false,
		},
		"#6 Different FirmwareVersion": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", "1.1.1"},
			false,
		},
		"#7 Component given as object, not pointer": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			true,
		},
		"#8 Component other than Ethernet given": {
			&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			FakeComponent{},
			false,
		},
	}
	for tn, tc := range cases {
		got := tc.eth.IsEqualTo(tc.comp)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}

}

func TestMemoryIsEqualTo(t *testing.T) {
	var cases = map[string]struct {
		mem  *Memory
		comp Component
		want bool
	}{
		"#0 All equal": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			true,
		},
		"#1 All different": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			&Memory{2, BaseObject{2}, "DIMM", 4096, 1333},
			false,
		},
		"#2 Different BaseObject.ID": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			&Memory{1, BaseObject{2}, "Samsung DDR3 DIMM", 16384, 1600},
			false,
		},
		"#3 Different ModelName": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			&Memory{1, BaseObject{1}, "DIMM", 16384, 1600},
			false,
		},
		"#4 Different Size": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 4096, 1600},
			false,
		},
		"#5 Different Speed": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1333},
			false,
		},
		"#6 Component given as object, not pointer": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			true,
		},
		"#7 Component other than Memory given": {
			&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			FakeComponent{},
			false,
		},
	}
	for tn, tc := range cases {
		got := tc.mem.IsEqualTo(tc.comp)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}

func TestFibreChannelCardIsEqualTo(t *testing.T) {
	var cases = map[string]struct {
		fc   *FibreChannelCard
		comp Component
		want bool
	}{
		"#0 All equal": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			true,
		},
		"#1 All different": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			&FibreChannelCard{2, BaseObject{2}, "Generic FC Card", "1 Gbit", "eeffeeffeeffeeff", "2.2.2"},
			false,
		},
		"#2 Different BaseObject.ID": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			&FibreChannelCard{1, BaseObject{2}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			false,
		},
		"#3 Different ModelName": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			&FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			false,
		},
		"#4 Different Speed": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "1 Gbit", "aabbccddeeff0011", "1.1.1"},
			false,
		},
		"#5 Different WWN": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "eeffeeffeeffeeff", "1.1.1"},
			false,
		},
		"#6 Different FirmwareVersion": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "2.2.2"},
			false,
		},
		"#7 Component given as object, not pointer": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			true,
		},
		"#8 Component other than FibreChannelCard given": {
			&FibreChannelCard{1, BaseObject{1}, "Saturn-X: LightPulse Fibre Channel Host Adapter", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			FakeComponent{},
			false,
		},
	}
	for tn, tc := range cases {
		got := tc.fc.IsEqualTo(tc.comp)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}

func TestMACIsInEths(t *testing.T) {
	var cases = map[string]struct {
		eths []*Ethernet
		mac  MACAddress
		want bool
	}{
		"#0 Contains": {
			[]*Ethernet{
				&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&Ethernet{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
			macs["aa:bb:cc:dd:ee:ff"],
			true,
		},
		"#1 Doesn't contain": {
			[]*Ethernet{
				&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&Ethernet{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
			macs["aa:aa:aa:aa:aa:aa"],
			false,
		},
	}
	for tn, tc := range cases {
		got := tc.mac.IsIn(tc.eths)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}

func TestGetBaseObject(t *testing.T) {
	var cases = map[string]struct {
		addr       Addr
		statusCode int
		json       string
		errMsg     string
		want       *BaseObject
	}{
		"#0 IP is assigned to a BaseObject": {
			"10.20.30.40",
			200,
			`{"count": 1, "results": [{"id": 1}]}`,
			"",
			&BaseObject{1},
		},
		"#1 IP is not assigned to any BaseObject": {
			"10.20.30.41",
			200,
			`{"count": 0, "results": []}`,
			"IP address 10.20.30.41 doesn't have any base objects",
			nil,
		},
		"#2 IP is assigned to >1 BaseObjects": {
			"10.20.30.41",
			200,
			`{"count": 2, "results": [{"id": 1}, {"id": 2}]}`,
			"IP address 10.20.30.41 has more than one base objects",
			nil,
		},
	}
	for tn, tc := range cases {
		server, client := MockServerClient(tc.statusCode, tc.json)
		defer server.Close()

		got, err := tc.addr.GetBaseObject(client)
		switch {
		case tc.errMsg != "":
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("%s\ndidn't get expected string: %q in err msg: %q", tn, tc.errMsg, err)
			}
		default:
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
			}
		}
	}
}

func TestCompareEthernets(t *testing.T) {
	var cases = map[string]struct {
		ethsOld []*Ethernet
		ethsNew []*Ethernet
		want    *Diff
	}{
		"#0 Empty diff": {
			[]*Ethernet{},
			[]*Ethernet{},
			&Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
		"#1 Create": {
			ethsOld: []*Ethernet{
				&Ethernet{
					ID:              1,
					BaseObject:      BaseObject{1},
					MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
					ModelName:       "Intel(R) Ethernet",
					Speed:           "1 Gbps",
					FirmwareVersion: "1.1.1",
				},
			},
			ethsNew: []*Ethernet{
				&Ethernet{
					ID:              1,
					BaseObject:      BaseObject{1},
					MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
					ModelName:       "Intel(R) Ethernet",
					Speed:           "1 Gbps",
					FirmwareVersion: "1.1.1",
				},
				&Ethernet{
					ID:              0,
					BaseObject:      BaseObject{1},
					MACAddress:      macs["aa:bb:cc:dd:ee:ff"],
					ModelName:       "Intel(R) Ethernet",
					Speed:           "10 Gbps",
					FirmwareVersion: "2.2.2",
				},
			},
			want: &Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:   0,
						Name: "Ethernet",
						Data: []byte(`{"id":0,"base_object":1,"mac":"aa:bb:cc:dd:ee:ff","model_name":"Intel(R) Ethernet","speed":4,"firmware_version":"2.2.2"}`),
						Component: &Ethernet{
							ID:              0,
							BaseObject:      BaseObject{1},
							MACAddress:      macs["aa:bb:cc:dd:ee:ff"],
							ModelName:       "Intel(R) Ethernet",
							Speed:           "10 Gbps",
							FirmwareVersion: "2.2.2",
						},
					},
				},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
		"#2 Update": {
			ethsOld: []*Ethernet{
				&Ethernet{
					ID:              1,
					BaseObject:      BaseObject{1},
					MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
					ModelName:       "",
					Speed:           "1 Gbps",
					FirmwareVersion: "1.1.1",
				},
			},
			ethsNew: []*Ethernet{
				&Ethernet{
					ID:              0,
					BaseObject:      BaseObject{1},
					MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
					ModelName:       "Intel(R) Ethernet",
					Speed:           "10 Gbps",
					FirmwareVersion: "2.2.2",
				},
			},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{
					&DiffComponent{
						ID:   1,
						Name: "Ethernet",
						Data: []byte(`{"id":1,"base_object":1,"mac":"a1:b2:c3:d4:e5:f6","model_name":"Intel(R) Ethernet","speed":4,"firmware_version":"2.2.2"}`),
						Component: &Ethernet{
							ID:              1,
							BaseObject:      BaseObject{1},
							MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
							ModelName:       "Intel(R) Ethernet",
							Speed:           "10 Gbps",
							FirmwareVersion: "2.2.2",
						},
					},
				},
				Delete: []*DiffComponent{},
			},
		},
		"#3 Delete": {
			ethsOld: []*Ethernet{
				&Ethernet{
					ID:              1,
					BaseObject:      BaseObject{1},
					MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
					ModelName:       "Intel(R) Ethernet",
					Speed:           "1 Gbps",
					FirmwareVersion: "1.1.1",
				},
			},
			ethsNew: []*Ethernet{},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:   1,
						Name: "Ethernet",
						Data: []byte(`{"id":1,"base_object":1,"mac":"a1:b2:c3:d4:e5:f6","model_name":"Intel(R) Ethernet","speed":3,"firmware_version":"1.1.1"}`),
						Component: &Ethernet{
							ID:              1,
							BaseObject:      BaseObject{1},
							MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
							ModelName:       "Intel(R) Ethernet",
							Speed:           "1 Gbps",
							FirmwareVersion: "1.1.1",
						},
					},
				},
			},
		},
	}
	for tn, tc := range cases {
		got, err := CompareEthernets(tc.ethsOld, tc.ethsNew)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(*got, *tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestCompareMemory(t *testing.T) {
	var cases = map[string]struct {
		memOld []*Memory
		memNew []*Memory
		want   *Diff
	}{
		"#0 Empty diff": {
			memOld: []*Memory{},
			memNew: []*Memory{},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
		"#1 Create": {
			memOld: []*Memory{},
			memNew: []*Memory{
				&Memory{0, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			},
			want: &Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:        0,
						Name:      "Memory",
						Data:      []byte(`{"base_object":1,"id":0,"model_name":"Samsung DDR3 DIMM","size":16384,"speed":1600}`),
						Component: &Memory{0, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
					},
				},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
		"#2 Delete (old > new && new > 0)": {
			memOld: []*Memory{
				&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
				&Memory{2, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			},
			memNew: []*Memory{
				&Memory{0, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "Memory",
						Data:      []byte(`{"base_object":1,"id":1,"model_name":"Samsung DDR3 DIMM","size":16384,"speed":1600}`),
						Component: &Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
					},
				},
			},
		},
		"#3 Delete (old > new && new == 0)": {
			memOld: []*Memory{
				&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			},
			memNew: []*Memory{},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "Memory",
						Data:      []byte(`{"base_object":1,"id":1,"model_name":"Samsung DDR3 DIMM","size":16384,"speed":1600}`),
						Component: &Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
					},
				},
			},
		},
		// Note that we test "Delete and Create" scenario instead of "Update" -
		// in case of Memory, the latter doesn't make sense because Memory
		// instances are not unique (in contrast to e.g. Ethernet, whose
		// instances can be distinguished by their MACAddresses).
		"#4 Update by \"Delete and Create\"": {
			memOld: []*Memory{
				&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			},
			memNew: []*Memory{
				&Memory{0, BaseObject{1}, "DIMM", 4096, 1333},
			},
			want: &Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:        0,
						Name:      "Memory",
						Data:      []byte(`{"base_object":1,"id":0,"model_name":"DIMM","size":4096,"speed":1333}`),
						Component: &Memory{0, BaseObject{1}, "DIMM", 4096, 1333},
					},
				},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "Memory",
						Data:      []byte(`{"base_object":1,"id":1,"model_name":"Samsung DDR3 DIMM","size":16384,"speed":1600}`),
						Component: &Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
					},
				},
			},
		},
		"#5 Don't do anything (both new and old Memory is the same)": {
			memOld: []*Memory{
				&Memory{1, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			},
			memNew: []*Memory{
				&Memory{0, BaseObject{1}, "Samsung DDR3 DIMM", 16384, 1600},
			},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
	}
	for tn, tc := range cases {
		got, err := CompareMemory(tc.memOld, tc.memNew)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(*got, *tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestCompareFibreChannelCards(t *testing.T) {
	var cases = map[string]struct {
		fccOld []*FibreChannelCard
		fccNew []*FibreChannelCard
		want   *Diff
	}{
		"#0 Empty diff": {
			fccOld: []*FibreChannelCard{},
			fccNew: []*FibreChannelCard{},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
		"#1 Create": {
			fccOld: []*FibreChannelCard{},
			fccNew: []*FibreChannelCard{
				&FibreChannelCard{0, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			},
			want: &Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:        0,
						Name:      "FibreChannelCard",
						Data:      []byte(`{"id":0,"base_object":1,"model_name":"Generic FC Card","speed":3,"wwn":"aabbccddeeff0011","firmware_version":"1.1.1"}`),
						Component: &FibreChannelCard{0, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
					},
				},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
		"#2 Delete (old > new && new > 0)": {
			fccOld: []*FibreChannelCard{
				&FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "1 Gbit", "", "1.1.1"},
				&FibreChannelCard{2, BaseObject{1}, "Generic FC Card", "1 Gbit", "", "1.1.1"},
				&FibreChannelCard{3, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "2.2.2"},
			},
			fccNew: []*FibreChannelCard{
				&FibreChannelCard{0, BaseObject{1}, "Generic FC Card", "1 Gbit", "", "1.1.1"},
			},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "FibreChannelCard",
						Data:      []byte(`{"id":1,"base_object":1,"model_name":"Generic FC Card","speed":1,"wwn":"","firmware_version":"1.1.1"}`),
						Component: &FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "1 Gbit", "", "1.1.1"},
					},
					&DiffComponent{
						ID:        3,
						Name:      "FibreChannelCard",
						Data:      []byte(`{"id":3,"base_object":1,"model_name":"Generic FC Card","speed":3,"wwn":"aabbccddeeff0011","firmware_version":"2.2.2"}`),
						Component: &FibreChannelCard{3, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "2.2.2"},
					},
				},
			},
		},
		"#3 Delete (old > new && new == 0)": {
			fccOld: []*FibreChannelCard{
				&FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			},
			fccNew: []*FibreChannelCard{},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "FibreChannelCard",
						Data:      []byte(`{"id":1,"base_object":1,"model_name":"Generic FC Card","speed":3,"wwn":"aabbccddeeff0011","firmware_version":"1.1.1"}`),
						Component: &FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
					},
				},
			},
		},
		// Note that we test "Delete and Create" scenario instead of "Update" -
		// in case of FibreChannelCard, the latter doesn't make sense because
		// FibreChannelCard instances are not guaranteed to be unique (and
		// that's because WWN field is not required).
		"#4 Update by \"Delete and Create\"": {
			fccOld: []*FibreChannelCard{
				&FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "4 Gbit", "", "1.1.1"},
			},
			fccNew: []*FibreChannelCard{
				&FibreChannelCard{0, BaseObject{1}, "Generic FC Card", "4 Gbit", "", "2.2.2"},
			},
			want: &Diff{
				Create: []*DiffComponent{
					&DiffComponent{
						ID:        0,
						Name:      "FibreChannelCard",
						Data:      []byte(`{"id":0,"base_object":1,"model_name":"Generic FC Card","speed":3,"wwn":"","firmware_version":"2.2.2"}`),
						Component: &FibreChannelCard{0, BaseObject{1}, "Generic FC Card", "4 Gbit", "", "2.2.2"},
					},
				},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{
					&DiffComponent{
						ID:        1,
						Name:      "FibreChannelCard",
						Data:      []byte(`{"id":1,"base_object":1,"model_name":"Generic FC Card","speed":3,"wwn":"","firmware_version":"1.1.1"}`),
						Component: &FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "4 Gbit", "", "1.1.1"},
					},
				},
			},
		},
		"#5 Don't do anything (both new and old FibreChannelCard is the same)": {
			fccOld: []*FibreChannelCard{
				&FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			},
			fccNew: []*FibreChannelCard{
				&FibreChannelCard{1, BaseObject{1}, "Generic FC Card", "4 Gbit", "aabbccddeeff0011", "1.1.1"},
			},
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
	}
	for tn, tc := range cases {
		got, err := CompareFibreChannelCards(tc.fccOld, tc.fccNew)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(*got, *tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestGetEthernets(t *testing.T) {
	var cases = []struct {
		file    string
		baseObj BaseObject
		want    []*Ethernet
	}{
		{
			"ethernet_components.json",
			BaseObject{1},
			[]*Ethernet{
				&Ethernet{
					ID:              2,
					BaseObject:      BaseObject{1},
					MACAddress:      macs["a1:b2:c3:d4:e5:f6"],
					ModelName:       "Intel(R) Ethernet 10G 4P X520/I350 rNDC",
					Speed:           "10 Gbps",
					FirmwareVersion: "1.1.1",
				},
			},
		},
	}

	for tn, tc := range cases {
		fixture, err := LoadFixture(ralphTestFixturesDir, tc.file)
		if err != nil {
			t.Fatalf("file: %s\n%s", tc.file, err)
		}
		server, client := MockServerClient(200, fixture)
		defer server.Close()

		got, err := tc.baseObj.GetEthernets(client)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("#%d\n%s", tn, err)
		}
	}
}

func TestGetMemory(t *testing.T) {
	var cases = []struct {
		file    string
		baseObj BaseObject
		want    []*Memory
	}{
		{
			"memory_components.json",
			BaseObject{1},
			[]*Memory{
				&Memory{
					ID:         2,
					BaseObject: BaseObject{1},
					ModelName:  "Samsung DDR3 DIMM",
					Size:       16384,
					Speed:      1600,
				},
			},
		},
	}

	for tn, tc := range cases {
		fixture, err := LoadFixture(ralphTestFixturesDir, tc.file)
		if err != nil {
			t.Fatalf("file: %s\n%s", tc.file, err)
		}
		server, client := MockServerClient(200, fixture)
		defer server.Close()

		got, err := tc.baseObj.GetMemory(client)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("#%d\n%s", tn, err)
		}
	}
}

func TestGetFibreChannelCard(t *testing.T) {
	var cases = []struct {
		file    string
		baseObj BaseObject
		want    []*FibreChannelCard
	}{
		{
			"fibre_channel_card_components.json",
			BaseObject{1},
			[]*FibreChannelCard{
				&FibreChannelCard{
					ID:              2,
					BaseObject:      BaseObject{1},
					ModelName:       "Saturn-X: LightPulse Fibre Channel Host Adapter",
					Speed:           "4 Gbit",
					WWN:             "aabbccddeeff0011",
					FirmwareVersion: "1.1.1",
				},
			},
		},
	}

	for tn, tc := range cases {
		fixture, err := LoadFixture(ralphTestFixturesDir, tc.file)
		if err != nil {
			t.Fatalf("file: %s\n%s", tc.file, err)
		}
		server, client := MockServerClient(200, fixture)
		defer server.Close()

		got, err := tc.baseObj.GetFibreChannelCards(client)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("#%d\n%s", tn, err)
		}
	}
}

func TestEthernetToString(t *testing.T) {
	ethernet := Ethernet{
		ID:              1,
		BaseObject:      BaseObject{1},
		MACAddress:      macs["aa:aa:aa:aa:aa:aa"],
		ModelName:       "Intel Corporation 82599EB 10-Gigabit SFI/SFP",
		Speed:           "10 Gbps",
		FirmwareVersion: "1.1.1",
	}
	want := `Ethernet{id: 1, base_object_id: 1, mac: aa:aa:aa:aa:aa:aa, model_name: Intel Corporation 82599EB 10-Gigabit SFI/SFP, speed: 10 Gbps, firmware_version: 1.1.1}`

	got := ethernet.String()
	if got != want {
		t.Errorf("\n got: %v\nwant: %v", got, want)
	}
}

func TestMemoryToString(t *testing.T) {
	memory := Memory{
		ID:         1,
		BaseObject: BaseObject{1},
		ModelName:  "Samsung DDR3 DIMM",
		Size:       16384,
		Speed:      1600,
	}
	want := `Memory{id: 1, base_object_id: 1, model_name: Samsung DDR3 DIMM, size: 16384, speed: 1600}`

	got := memory.String()
	if got != want {
		t.Errorf("\n got: %v\nwant: %v", got, want)
	}
}

func TestFibreChannelCardToString(t *testing.T) {
	memory := FibreChannelCard{
		ID:              1,
		BaseObject:      BaseObject{1},
		ModelName:       "Saturn-X: LightPulse Fibre Channel Host Adapter",
		Speed:           "4 Gbit",
		WWN:             "aabbccddeeff0011",
		FirmwareVersion: "1.1.1",
	}
	want := `FibreChannelCard{id: 1, base_object_id: 1, model_name: Saturn-X: LightPulse Fibre Channel Host Adapter, speed: 4 Gbit, wwn: aabbccddeeff0011, firmware_version: 1.1.1}`

	got := memory.String()
	if got != want {
		t.Errorf("\n got: %v\nwant: %v", got, want)
	}
}

func TestEthSpeedMarshalJSON(t *testing.T) {
	var cases = []struct {
		speed EthSpeed
		want  []byte
	}{
		{"10 Mbps", []byte(strconv.Itoa(ethSpeedChoices["10 Mbps"]))},
		{"100 Mbps", []byte(strconv.Itoa(ethSpeedChoices["100 Mbps"]))},
		{"1 Gbps", []byte(strconv.Itoa(ethSpeedChoices["1 Gbps"]))},
		{"10 Gbps", []byte(strconv.Itoa(ethSpeedChoices["10 Gbps"]))},
		{"40 Gbps", []byte(strconv.Itoa(ethSpeedChoices["40 Gbps"]))},
		{"100 Gbps", []byte(strconv.Itoa(ethSpeedChoices["100 Gbps"]))},
		{"unknown speed", []byte(strconv.Itoa(ethSpeedChoices["unknown speed"]))},
	}
	for _, tc := range cases {
		got, err := tc.speed.MarshalJSON()
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !TestEqByte(got, tc.want) {
			t.Errorf("\n got: %v\nwant: %v", got, tc.want)
		}
	}
}

func TestFCCSpeedMarshalJSON(t *testing.T) {
	var cases = []struct {
		speed FCCSpeed
		want  []byte
	}{
		{"1 Gbit", []byte(strconv.Itoa(fccSpeedChoices["1 Gbit"]))},
		{"2 Gbit", []byte(strconv.Itoa(fccSpeedChoices["2 Gbit"]))},
		{"4 Gbit", []byte(strconv.Itoa(fccSpeedChoices["4 Gbit"]))},
		{"8 Gbit", []byte(strconv.Itoa(fccSpeedChoices["8 Gbit"]))},
		{"16 Gbit", []byte(strconv.Itoa(fccSpeedChoices["16 Gbit"]))},
		{"32 Gbit", []byte(strconv.Itoa(fccSpeedChoices["32 Gbit"]))},
		{"unknown speed", []byte(strconv.Itoa(fccSpeedChoices["unknown speed"]))},
	}
	for _, tc := range cases {
		got, err := tc.speed.MarshalJSON()
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !TestEqByte(got, tc.want) {
			t.Errorf("\n got: %v\nwant: %v", got, tc.want)
		}
	}
}
