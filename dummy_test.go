package main

import "testing"

func TestHello(t *testing.T) {
	expected := "Hell-o world!"
	actual := "Hell-o world!"
	if actual != expected {
		t.Error("Test failed")
	}
}
