package builder_test

import (
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
)

func TestBuilder(t *testing.T) {
	if api.TestString() != "this is a string" {
		t.Fatal("Wrong test string :")
	}
}
