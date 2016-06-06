package main

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/juju/testing/checkers"
)

func TestNewClient(t *testing.T) {
	var cases = map[string]struct {
		scannedAddr Addr
		ralphURL    string
		apiKey      string
		errMsg      string
		want        *Client
	}{
		"#0 All params provided are correct": {
			Addr("10.20.30.40"),
			"http://localhost:8080/api",
			"abcdefghijklmnopqrstuwxyz0123456789ABCDE",
			"",
			&Client{
				"10.20.30.40",
				"http://localhost:8080/api",
				"abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				"", // apiVersion
				&http.Client{Timeout: time.Second * 10},
			},
		},
		"#1 Missing API key": {
			Addr("10.20.30.40"),
			"http://localhost:8080/api",
			"",
			"API key is missing",
			nil,
		},
		"#2 Missing Ralph URL": {
			Addr("10.20.30.40"),
			"",
			"abcdefghijklmnopqrstuwxyz0123456789ABCDE",
			"Ralph's URL is missing",
			nil,
		},
	}
	for tn, tc := range cases {
		got, err := NewClient(tc.ralphURL, tc.apiKey, tc.scannedAddr, nil)
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

func TestSendToRalph(t *testing.T) {
	var cases = map[string]struct {
		method     string
		endpoint   string
		data       []byte
		statusCode int
		errMsg     string
		want       int
	}{
		"#0 Ralph responds with >299": {
			"POST",
			"non-existing-endpoint",
			[]byte{},
			404,
			"error while sending to",
			404,
		},
		// Other cases are covered in TestSendDiffToRalph
	}

	for tn, tc := range cases {
		server, client := MockServerClient(tc.statusCode, `{}`)
		defer server.Close()

		got, err := client.SendToRalph(tc.method, tc.endpoint, tc.data)
		switch {
		case tc.errMsg != "":
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("%s\ndidn't get expected string: %q in err msg: %q", tn, tc.errMsg, err)
			}
			if got != tc.want {
				t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
			}
		default:
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if got != tc.want {
				t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
			}
		}
	}

}
