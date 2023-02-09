package service_test

import "testing"

func TestService(t *testing.T) {
	if service.TestString() != "this is a string" {
		t.Fatal(":")
	}
}