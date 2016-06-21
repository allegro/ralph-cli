package main

import (
	"testing"
)

func TestValidationError(t *testing.T) {

	err1 := "something went wrong #1"
	err2 := "something went wrong #2"
	err3 := "something went wrong #3"

	var cases = map[string]struct {
		path    string
		errMsgs []*string
		want    string
	}{
		"#0 No errors": {
			"/some/file/being/validated",
			[]*string{},
			"no errors",
		},
		"#1 Single error": {
			"/some/file/being/validated",
			[]*string{&err1},
			"validation error in /some/file/being/validated: something went wrong #1",
		},
		"#2 Multiple errors": {
			"/some/file/being/validated",
			[]*string{&err1, &err2, &err3},
			"validation errors in /some/file/being/validated (3 in total): (1) something went wrong #1; (2) something went wrong #2; (3) something went wrong #3",
		},
	}

	for tn, tc := range cases {
		got := NewValidationError(tc.path, tc.errMsgs).Error()
		if got != tc.want {
			t.Errorf("%s\n got: %q\nwant: %q", tn, got, tc.want)
		}
	}

}
