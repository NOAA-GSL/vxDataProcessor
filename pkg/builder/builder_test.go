package builder_test

import (
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"testing"
	"fmt"
)

var gp  = builder.GoodnessPolarity(1)
var minorThreshold = builder.Threshold(0.05)
var majorThreshold = builder.Threshold(0.01)
var cellPtr = builder.GetBuilder("TwoSampleTTest")

func TestTwoSampleTTestBuilder_test_identical(t *testing.T) {
	cellPtr.SetGoodnessPolarity(gp)
	cellPtr.SetMinorThreshold(minorThreshold)
	cellPtr.SetMajorThreshold(majorThreshold)
	cellPtr.SetInputData(builder.DerivedDataElement{
			CtlPop: []float64{0.2,1.3,3.2,4.5},
			ExpPop: []float64{0.2,1.3,3.2,4.5},
	})
	cellPtr.ComputeSignificance(cellPtr.Data)
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 0 {
		t.Fatal("test_indentical_wrong value :", cellPtr.Value)
	}
}
func TestTwoSampleTTestBuilder_test_2(t *testing.T) {
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{1.0,1.1,1.2,1.09,1.08},
		ExpPop: []float64{0.99,1.09,1.19,1.08,1.07},
	})
	cellPtr.ComputeSignificance(cellPtr.Data)
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 0 {
		t.Fatal("test_2_wrong value :", cellPtr.Value)
	}
}
func TestTwoSampleTTestBuilder_test_1(t *testing.T) {
		cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2,0.31,0.11,0},
		ExpPop: []float64{0.2,0.3,0.1,0.9},
	})
	cellPtr.ComputeSignificance(cellPtr.Data)
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 0 {
		t.Fatal("test_1 wrong value :", cellPtr.Value)
	}
}
func TestTwoSampleTTestBuilder_test_0(t *testing.T) {
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2,0.303,0.101,0},
		ExpPop: []float64{0.2,0.3,0.1,0.01},
	})
	cellPtr.ComputeSignificance(cellPtr.Data)
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 0 {
		t.Fatal("test_0 wrong value :", cellPtr.Value)
	}
}

func TestTwoSampleTTestBuilder_different_lengths(t *testing.T) {
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2,0.303,0.101,0},
		ExpPop: []float64{0.2,0.3,0.1},
	})
	cellPtr.ComputeSignificance(cellPtr.Data)
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 0 {
		t.Fatal("Wrong value :", cellPtr.Value)
	}
}
