package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseComponents(t *testing.T) {
	var cases = map[string]struct {
		components string
		errMsg     string
		want       *map[string]bool
	}{
		"#0 All valid components": {
			"eth,mem,fcc,cpu,disk",
			"",
			&map[string]bool{
				"none": false,
				"all":  false,
				"eth":  true,
				"mem":  true,
				"fcc":  true,
				"cpu":  true,
				"disk": true,
			},
		},
		"#1 Unknown component": {
			"eth,mem,printer",
			"unknown component: printer",
			nil,
		},
		"#2 \"none\" option should be used exclusively": {
			"none,eth,mem",
			"invalid combination",
			nil,
		},
		"#3 \"all\" option should be used exclusively": {
			"all,eth,mem",
			"invalid combination",
			nil,
		},
	}

	for tn, tc := range cases {
		got, err := parseComponents(tc.components)
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
