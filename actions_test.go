package main

import (
	"testing"

	"github.com/juju/testing/checkers"
)

var actionsTestFixturesDir = "./actions_test_fixtures"

func TestExcludeMgmt(t *testing.T) {
	var cases = []struct {
		file string
		eths []*Ethernet
		ip   Addr
		want []*Ethernet
	}{
		{
			"exclude_mgmt.json",
			[]*Ethernet{
				&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&Ethernet{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
				&Ethernet{3, BaseObject{3}, macs["74:86:7a:ee:20:e8"], "", "", ""},
			},
			"10.20.30.40",
			[]*Ethernet{
				&Ethernet{1, BaseObject{1}, macs["aa:bb:cc:dd:ee:ff"], "", "", ""},
				&Ethernet{2, BaseObject{2}, macs["a1:b2:c3:d4:e5:f6"], "", "", ""},
			},
		},
	}

	for tn, tc := range cases {
		fixture, err := LoadFixture(actionsTestFixturesDir, tc.file)
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

func TestExcludeExposedInDHCP(t *testing.T) {
	var cases = map[string]struct {
		diff       *Diff
		statusCode int
		file       string
		want       *Diff
	}{
		"#0 Not exposed in DHCP shouldn't be excluded": {
			diff: &Diff{
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
			statusCode: 200,
			file:       "exclude_exposed_in_dhcp_false.json",
			want: &Diff{
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
		},
		"#1 Exposed in DHCP should be excluded": {
			diff: &Diff{
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
			statusCode: 200,
			file:       "exclude_exposed_in_dhcp_true.json",
			want: &Diff{
				Create: []*DiffComponent{},
				Update: []*DiffComponent{},
				Delete: []*DiffComponent{},
			},
		},
	}

	for tn, tc := range cases {
		fixture, err := LoadFixture(actionsTestFixturesDir, tc.file)
		if err != nil {
			t.Fatalf("file: %s\n%s", tc.file, err)
		}
		server, client := MockServerClient(tc.statusCode, fixture)
		defer server.Close()

		got, err := ExcludeExposedInDHCP(tc.diff, client, true)
		if err != nil {
			t.Fatalf("%s\nerr: %s", tn, err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("%s\n%s", tn, err)
		}
	}
}
