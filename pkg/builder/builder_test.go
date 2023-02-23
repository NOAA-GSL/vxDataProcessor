package builder_test

import (
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"testing"
)

func TestBuilder(t *testing.T) {
	if builder.TestString() != "this is a string from builder" {
		t.Fatal("Wrong test string :")
	}
}
