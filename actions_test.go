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

func TestUpdateBIOSAndFirmwareVersions(t *testing.T) {
	var cases = map[string]struct {
		scanResult *ScanResult
		dcAsset    *DataCenterAsset
		want       bool
		// We may also check if unchanged dcAsset's fields are getting nil.
	}{
		"#0 Different FirmwareVersion": {
			&ScanResult{FirmwareVersion: "2.2.2"},
			&DataCenterAsset{FirmwareVersion: PtrToStr("1.1.1")},
			true,
		},
		"#1 Different FirmwareVersion (dcAsset.FirmwareVersion == nil)": {
			&ScanResult{FirmwareVersion: "1.1.1"},
			&DataCenterAsset{},
			true,
		},
		"#2 Equal FirmwareVersion": {
			&ScanResult{FirmwareVersion: "1.1.1"},
			&DataCenterAsset{FirmwareVersion: PtrToStr("1.1.1")},
			false,
		},
		"#3 Different BIOSVersion": {
			&ScanResult{BIOSVersion: "2.2.2"},
			&DataCenterAsset{BIOSVersion: PtrToStr("1.1.1")},
			true,
		},
		"#4 Different BIOSVersion (dcAsset.BIOSVersion == nil)": {
			&ScanResult{BIOSVersion: "1.1.1"},
			&DataCenterAsset{},
			true,
		},
		"#5 Equal BIOSVersion": {
			&ScanResult{BIOSVersion: "1.1.1"},
			&DataCenterAsset{BIOSVersion: PtrToStr("1.1.1")},
			false,
		},
	}
	for tn, tc := range cases {
		got := updateBIOSAndFirmwareVersions(tc.scanResult, tc.dcAsset)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}

func TestUpdateModelName(t *testing.T) {
	var cases = map[string]struct {
		scanResult  *ScanResult
		dcAsset     *DataCenterAsset
		wantChanged bool
		wantRemarks string
	}{
		"#0 Different": {
			&ScanResult{ModelName: "Dell PowerEdge R620"},
			&DataCenterAsset{Remarks: PtrToStr(">>> ralph-cli: detected model name: Dell PowerEdge R720 <<<")},
			true,
			">>> ralph-cli: detected model name: Dell PowerEdge R620 <<<",
		},
		"#1 Different (empty remarks)": {
			&ScanResult{ModelName: "Dell PowerEdge R620"},
			&DataCenterAsset{Remarks: PtrToStr("")},
			true,
			">>> ralph-cli: detected model name: Dell PowerEdge R620 <<<",
		},
		"#2 Different (empty ScanResult.ModelName)": {
			&ScanResult{ModelName: ""},
			&DataCenterAsset{Remarks: PtrToStr("some remark")},
			false,
			"some remark",
		},
		"#3 Equal": {
			&ScanResult{ModelName: "Dell PowerEdge R620"},
			&DataCenterAsset{Remarks: PtrToStr(">>> ralph-cli: detected model name: Dell PowerEdge R620 <<<")},
			false,
			">>> ralph-cli: detected model name: Dell PowerEdge R620 <<<",
		},
		"#4 Remark with ModelName is appended non-destructively": {
			&ScanResult{ModelName: "Dell PowerEdge R620"},
			&DataCenterAsset{Remarks: PtrToStr("some remark")},
			true,
			"some remark\n>>> ralph-cli: detected model name: Dell PowerEdge R620 <<<",
		},
		"#5 Different (ModelName in remarks, but not in ScanResult)": {
			&ScanResult{ModelName: ""},
			&DataCenterAsset{Remarks: PtrToStr(">>> ralph-cli: detected model name: Dell PowerEdge R620 <<<")},
			true,
			"",
		},
	}
	for tn, tc := range cases {
		got := updateModelName(tc.scanResult, tc.dcAsset)
		if got != tc.wantChanged {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.wantChanged)
		}
		// Since dcAsset.Remarks is mutated by updateModelName, we need to check
		// it as well.
		if *tc.dcAsset.Remarks != tc.wantRemarks {
			t.Errorf("%s\n got remarks: %v\nwant remarks: %v", tn, *tc.dcAsset.Remarks, tc.wantRemarks)
		}
	}
}

func TestVerifySerialNumber(t *testing.T) {
	var cases = map[string]struct {
		dcAsset    *DataCenterAsset
		scanResult *ScanResult
		want       bool
	}{
		"#0 Different": {
			&DataCenterAsset{SerialNumber: PtrToStr("SN1234")},
			&ScanResult{SN: "SN4321"},
			true,
		},
		"#1 Different (dcAsset.SerialNumber == nil)": {
			&DataCenterAsset{},
			&ScanResult{SN: "SN4321"},
			true,
		},
		"#2 Equal": {
			&DataCenterAsset{SerialNumber: PtrToStr("SN1234")},
			&ScanResult{SN: "SN1234"},
			false,
		},
	}
	for tn, tc := range cases {
		got := verifySerialNumber(tc.dcAsset, tc.scanResult, true)
		if got != tc.want {
			t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
		}
	}
}
