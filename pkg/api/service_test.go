package service_test

import (
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
)

func TestService(t *testing.T) {
	if service.TestString() != "this is a string" {
		t.Fatal("Wrong test string :")
	}
}