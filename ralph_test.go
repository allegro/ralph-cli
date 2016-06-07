package main

import (
	"reflect"
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

func TestNewEthernetComponent(t *testing.T) {
	var cases = map[string]struct {
		mac     MACAddress
		baseObj *BaseObject
		speed   string
		want    *EthernetComponent
	}{
		"#0 No speed provided": {
			macs["aa:bb:cc:dd:ee:ff"],
			&BaseObject{1},
			"",
			&EthernetComponent{0, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "unknown speed", ""},
		},
	}
	for tn, tc := range cases {
		got := NewEthernetComponent(tc.mac, tc.baseObj, tc.speed)
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestIsEqualTo(t *testing.T) {
	var cases = map[string]struct {
		ec1  *EthernetComponent
		ec2  *EthernetComponent
		want bool
	}{
		"#0 All equal": {
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			true,
		},
		"#1 All different": {
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&EthernetComponent{2, BaseObject{2}, macs["aa:aa:aa:aa:aa:aa"], "eth0", "10 Gbps", "Intel Corporation 82599EB 10-Gigabit SFI/SFP"},
			false,
		},
		"#2 Different BaseObject.ID": {
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&EthernetComponent{1, BaseObject{2}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			false,
		},
		"#3 Different MACAddress": {
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&EthernetComponent{1, BaseObject{1}, macs["aa:aa:aa:aa:aa:aa"], "", "", ""},
			false,
		},
		"#4 Different Label": {
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "eth0", "", ""},
			false,
		},
		"#5 Different Speed": {
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "10 Gbps", ""},
			false,
		},
		"#6 Different Model": {
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", "Intel Corporation 82599EB 10-Gigabit SFI/SFP"},
			false,
		},
	}
	for tn, tc := range cases {
		got := tc.ec1.IsEqualTo(tc.ec2)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}

}

func TestContains(t *testing.T) {
	var cases = map[string]struct {
		eths []*EthernetComponent
		mac  MACAddress
		want bool
	}{
		"#0 Contains": {
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&EthernetComponent{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
			macs["aa:bb:cc:dd:ee:ff"],
			true,
		},
		"#1 Doesn't contain": {
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&EthernetComponent{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
			macs["aa:aa:aa:aa:aa:aa"],
			false,
		},
	}
	for tn, tc := range cases {
		got := contains(tc.eths, tc.mac)
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

func TestExcludeMgmt(t *testing.T) {
	var cases = []struct {
		file string
		eths []*EthernetComponent
		ip   Addr
		want []*EthernetComponent
	}{
		{
			"exclude_mgmt.json",
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&EthernetComponent{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
				&EthernetComponent{3, BaseObject{3}, macs["74:86:7a:ee:20:e8"], "", "", ""},
			},
			"10.20.30.40",
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&EthernetComponent{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
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

		got, err := ExcludeMgmt(tc.eths, tc.ip, client)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("#%d\n%s", tn, err)
		}

	}
}

func TestIsEmpty(t *testing.T) {
	var cases = map[string]struct {
		diff *DiffEthernetComponent
		want bool
	}{
		"#0 Is empty": {
			&DiffEthernetComponent{
				Create: []*EthernetComponent{},
				Update: []*EthernetComponent{},
				Delete: []*EthernetComponent{},
			},
			true,
		},
		"#1 Is not empty": {
			&DiffEthernetComponent{
				Create: []*EthernetComponent{
					&EthernetComponent{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
				},
				Update: []*EthernetComponent{},
				Delete: []*EthernetComponent{},
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

func TestCompareEthernetComponents(t *testing.T) {
	var cases = map[string]struct {
		ethsOld []*EthernetComponent
		ethsNew []*EthernetComponent
		want    *DiffEthernetComponent
	}{
		"#0 Empty diff": {
			[]*EthernetComponent{},
			[]*EthernetComponent{},
			&DiffEthernetComponent{Create: []*EthernetComponent{}, Update: []*EthernetComponent{}, Delete: []*EthernetComponent{}},
		},
		"#1 Create": {
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
				&EthernetComponent{0, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
			},
			&DiffEthernetComponent{
				Create: []*EthernetComponent{
					&EthernetComponent{0, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				},
				Update: []*EthernetComponent{},
				Delete: []*EthernetComponent{},
			},
		},
		"#2 Update": {
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
			[]*EthernetComponent{
				&EthernetComponent{0, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "eth0", "10 Gbps", ""},
			},
			&DiffEthernetComponent{
				Create: []*EthernetComponent{},
				Update: []*EthernetComponent{
					&EthernetComponent{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "eth0", "10 Gbps", ""},
				},
				Delete: []*EthernetComponent{},
			},
		},
		"#3 Delete": {
			[]*EthernetComponent{
				&EthernetComponent{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
			[]*EthernetComponent{},
			&DiffEthernetComponent{
				Create: []*EthernetComponent{},
				Update: []*EthernetComponent{},
				Delete: []*EthernetComponent{
					&EthernetComponent{1, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
				},
			},
		},
	}
	for tn, tc := range cases {
		got, err := CompareEthernetComponents(tc.ethsOld, tc.ethsNew)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(*got, *tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}

func TestGetEthernetComponents(t *testing.T) {
	var cases = []struct {
		file    string
		baseObj BaseObject
		want    []*EthernetComponent
	}{
		{
			"ethernet_components.json",
			BaseObject{1},
			[]*EthernetComponent{
				&EthernetComponent{2, BaseObject{1}, macs["a1:b2:c3:d4:e5:f6"], "eth0", "10 Gbps", ""},
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

		got, err := tc.baseObj.GetEthernetComponents(client)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("#%d\n%s", tn, err)
		}
	}
}

// TODO(xor-xor): Refactor SendDiffToRalph (or better yet, MockServerClient) to support
// testing scenarios, where multiple, different HTTP status codes are returned.
func TestSendDiffToRalph(t *testing.T) {
	var cases = map[string]struct {
		diff       *DiffEthernetComponent
		dryRun     bool
		statusCode int
		want       []int
	}{
		"#0 Empty diff": {
			&DiffEthernetComponent{
				Create: []*EthernetComponent{},
				Update: []*EthernetComponent{},
				Delete: []*EthernetComponent{},
			},
			false,
			0, // In this case, statusCode doesn't really matter.
			[]int{},
		},
		"#1 Dry run": {
			&DiffEthernetComponent{
				Create: []*EthernetComponent{
					&EthernetComponent{0, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				},
				Update: []*EthernetComponent{
					&EthernetComponent{1, BaseObject{2}, macs["aa:aa:aa:aa:aa:aa"], "", "", ""},
				},
				Delete: []*EthernetComponent{
					&EthernetComponent{2, BaseObject{3}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
				},
			},
			true,
			0, // In this case, statusCode doesn't really matter.
			[]int{},
		},
		"#2 Create": {
			&DiffEthernetComponent{
				Create: []*EthernetComponent{
					&EthernetComponent{0, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				},
				Update: []*EthernetComponent{},
				Delete: []*EthernetComponent{},
			},
			false,
			201,
			[]int{201},
		},
		"#3 Update": {
			&DiffEthernetComponent{
				Create: []*EthernetComponent{},
				Update: []*EthernetComponent{
					&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				},
				Delete: []*EthernetComponent{},
			},
			false,
			200,
			[]int{200},
		},
		"#4 Delete": {
			&DiffEthernetComponent{
				Create: []*EthernetComponent{},
				Update: []*EthernetComponent{},
				Delete: []*EthernetComponent{
					&EthernetComponent{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
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
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}
