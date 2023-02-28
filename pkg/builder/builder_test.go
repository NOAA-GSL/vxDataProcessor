package builder_test

import (
	"fmt"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"testing"
)

var gp = builder.GoodnessPolarity(1)
var minorThreshold = builder.Threshold(0.05)
var majorThreshold = builder.Threshold(0.01)
var cellPtr = builder.GetBuilder("TwoSampleTTest")

// test for zero variance
func TestTwoSampleTTestBuilder_test_identical(t *testing.T) {
	cellPtr.SetGoodnessPolarity(gp)
	cellPtr.SetMinorThreshold(minorThreshold)
	cellPtr.SetMajorThreshold(majorThreshold)
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2, 1.3, 3.2, 4.5},
		ExpPop: []float64{0.2, 1.3, 3.2, 4.5},
	})
	err := cellPtr.ComputeSignificance(cellPtr.Data)
	if err == nil {
		t.Fatal("TestTwoSampleTTestBuilder_test_identical - no error message : should be 'sample has zero variance'")
	}
}
func TestTwoSampleTTestBuilder_test_2(t *testing.T) {
	cellPtr.SetGoodnessPolarity(gp)
	cellPtr.SetMinorThreshold(minorThreshold)
	cellPtr.SetMajorThreshold(majorThreshold)
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{1.0, 1.1, 1.2, 1.15, 1.09},
		ExpPop: []float64{0.9, 1.0, 1.1, 1.08, 1.05},
	})
	err := cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - error message : ", err))
	}
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 2 {
		t.Fatal("test_2_wrong value :", cellPtr.Value)
	}
}
func TestTwoSampleTTestBuilder_test_1(t *testing.T) {
	cellPtr.SetGoodnessPolarity(gp)
	cellPtr.SetMinorThreshold(minorThreshold)
	cellPtr.SetMajorThreshold(majorThreshold)
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{1.0, 1.2, 1.3, 1.15, 1.09},
		ExpPop: []float64{0.9, 1.0, 1.1, 1.08, 1.05},
	})
	err := cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - error message : ", err))
	}

	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 1 {
		t.Fatal("test_1 wrong value :", cellPtr.Value)
	}
}
func TestTwoSampleTTestBuilder_test_0(t *testing.T) {
	cellPtr.SetGoodnessPolarity(gp)
	cellPtr.SetMinorThreshold(minorThreshold)
	cellPtr.SetMajorThreshold(majorThreshold)
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2, 0.303, 0.101, 0},
		ExpPop: []float64{0.2, 0.3, 0.1, 0.01},
	})
	err := cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - error message : ", err))
	}
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 0 {
		t.Fatal("test_0 wrong value :", cellPtr.Value)
	}
}

func TestTwoSampleTTestBuilder_different_lengths(t *testing.T) {
	cellPtr.SetGoodnessPolarity(gp)
	cellPtr.SetMinorThreshold(minorThreshold)
	cellPtr.SetMajorThreshold(majorThreshold)
	cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2, 0.303, 0.101, 0},
		ExpPop: []float64{0.2, 0.3, 0.1},
	})
	err := cellPtr.ComputeSignificance(cellPtr.Data)
	if err == nil {
		t.Fatal("TestTwoSampleTTestBuilder_test_identical - no error message : should be 'sample has zero variance'")
	}
}
