package cmd

import "testing"

func TestHello(t *testing.T) {
	expected := "Hello Go!"
	actual := "Hello Go!"
	if actual != expected {
		t.Error("Test failed")
	}
}
