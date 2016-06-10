package main

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/juju/testing/checkers"
)

func TestNewClient(t *testing.T) {
	var cases = []struct {
		config      *Config
		scannedAddr Addr
		errMsg      string
		want        *Client
	}{
		{
			&Config{
				RalphAPIURL:   "http://localhost:8080/api",
				RalphAPIKey:   "abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				ClientTimeout: 10,
			},
			Addr("10.20.30.40"),
			"",
			&Client{
				"10.20.30.40",
				"http://localhost:8080/api",
				"abcdefghijklmnopqrstuwxyz0123456789ABCDE",
				"", // apiVersion
				&http.Client{Timeout: time.Second * 10},
			},
		},
	}
	for tn, tc := range cases {
		got, err := NewClient(tc.config, tc.scannedAddr, nil)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if eq, err := checkers.DeepEqual(got, tc.want); !eq {
			t.Errorf("#%d\n%s", tn, err)
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
		"#0 Ralph responds with >= 400": {
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

func TestGetFromRalph(t *testing.T) {
	var cases = map[string]struct {
		endpoint   string
		query      string
		statusCode int
		errMsg     string
		want       []byte
	}{
		"#0 Ralph responds with >= 400": {
			"non-existing-endpoint",
			"some_valid_query",
			404,
			"error while sending a GET request to Ralph",
			[]byte{},
		},
	}

	for tn, tc := range cases {
		server, client := MockServerClient(tc.statusCode, `{}`)
		defer server.Close()

		got, err := client.GetFromRalph(tc.endpoint, tc.query)
		switch {
		case tc.errMsg != "":
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("%s\ndidn't get expected string: %q in err msg: %q", tn, tc.errMsg, err)
			}
			if !TestEqByte(got, tc.want) {
				t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
			}
		default:
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if !TestEqByte(got, tc.want) {
				t.Errorf("%s\n got: %v\nwant: %v", tn, got, tc.want)
			}
		}
	}
}
