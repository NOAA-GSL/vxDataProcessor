package builder_test

import (
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
)

func TestBuilder(t *testing.T) {
	if builder.TestString() != "this is a string from builder" {
		t.Fatal("Wrong test string :")
	}
}
