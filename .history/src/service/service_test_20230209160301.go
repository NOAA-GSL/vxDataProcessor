package service_test

import (
	"testing"

	"github.com/NOAA-GSL/vxGoDataProcessing/src/service"
)

func TestService(t *testing.T) {
	if service.TestString() != "this a string" {
		t.Fatal("Wrong test string :")
	}
}