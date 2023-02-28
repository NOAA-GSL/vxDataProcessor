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
	err := cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetGoodnessPolarity - error message : ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2, 1.3, 3.2, 4.5},
		ExpPop: []float64{0.2, 1.3, 3.2, 4.5},
	})
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err == nil {
		t.Fatal("TestTwoSampleTTestBuilder_test_identical - ComputeSignificance - no error message : should be 'sample has zero variance'")
	}
}
func TestTwoSampleTTestBuilder_test_2(t *testing.T) {
	err := cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetGoodnessPolarity - error message : ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{1.0, 1.1, 1.2, 1.15, 1.09},
		ExpPop: []float64{0.9, 1.0, 1.1, 1.08, 1.05},
	})
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 2 {
		t.Fatal("test_2_wrong value :", cellPtr.Value)
	}
}
func TestTwoSampleTTestBuilder_test_1(t *testing.T) {
	err := cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetGoodnessPolarity - error message : ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{1.0, 1.2, 1.3, 1.15, 1.09},
		ExpPop: []float64{0.9, 1.0, 1.1, 1.08, 1.05},
	})
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}

	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 1 {
		t.Fatal("test_1 wrong value :", cellPtr.Value)
	}
}
func TestTwoSampleTTestBuilder_test_0(t *testing.T) {
	err := cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetGoodnessPolarity - error message : ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2, 0.303, 0.101, 0},
		ExpPop: []float64{0.2, 0.3, 0.1, 0.01},
	})
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - ComputeSignificance - error message : ", err))
	}
	fmt.Println("Pval is", cellPtr.StatValue, "value is ", cellPtr.Value)
	if cellPtr.Value != 0 {
		t.Fatal("test_0 wrong value :", cellPtr.Value)
	}
}

func TestTwoSampleTTestBuilder_different_lengths(t *testing.T) {
	err := cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetGoodnessPolarity - error message : ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(builder.DerivedDataElement{
		CtlPop: []float64{0.2, 0.303, 0.101, 0},
		ExpPop: []float64{0.2, 0.3, 0.1},
	})
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err == nil {
		t.Fatal("TestTwoSampleTTestBuilder_test_identical - ComputeSignificance - no error message : should be 'sample has zero variance'")
	}
}
