package builder_test

import (
	"fmt"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"testing"
	"time"
)

var gp = builder.GoodnessPolarity(1)
var minorThreshold = builder.Threshold(0.05)
var majorThreshold = builder.Threshold(0.01)
var cellPtr = builder.GetBuilder("TwoSampleTTest")

// test for zero variance
func TestTwoSampleTTestBuilder_test_identical(t *testing.T) {
	err := (*cellPtr).SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetGoodnessPolarity - error message : ", err))
	}
	err = (*cellPtr).SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetMinorThreshold - error message : ", err))
	}
	err = (*cellPtr).SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetMajorThreshold - error message : ", err))
	}
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_identical - SetInputData - error message : ", err))
	}

	var epoch = time.Now().Unix()
	var ctlData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.1,
			Time:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.1,
			Time:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResultPtr *builder.QueryResult = new(builder.QueryResult)
	queryResultPtr.CtlData = &ctlData
	queryResultPtr.ExpData = &expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResultPtr, statistic, "PreCalcRecord")
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}

	err = (*cellPtr).ComputeSignificance()
	if err == nil {
		t.Fatal("TestTwoSampleTTestBuilder_test_identical - ComputeSignificance - no error message : should be 'sample has zero variance'")
	}
}

// this test has inputs that should return a value of 2
func TestTwoSampleTTestBuilder_test_2(t *testing.T) {
	var cellPtr = builder.NewTwoSampleTTestBuilder()
	var value = new(int)
	(*cellPtr).SetValuePtr(*value)
	err := (*cellPtr).SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetGoodnessPolarity - error message : ", err))
	}
	err = (*cellPtr).SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetMinorThreshold - error message : ", err))
	}
	err = (*cellPtr).SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetMajorThreshold - error message : ", err))
	}
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - SetInputData - error message : ", err))
	}
	var epoch = time.Now().Unix()
	var ctlData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.01,
			Time:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.2,
			Time:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResultPtr *builder.QueryResult = new(builder.QueryResult)
	queryResultPtr.CtlData = &ctlData
	queryResultPtr.ExpData = &expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResultPtr, statistic, "PreCalcRecord")
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - ComputeSignificance - error message : ", err))
	}

	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_2 - ComputeSignificance - error message : ", err))
	}
	if *(*cellPtr).ValuePtr != 2 {
		t.Fatal("test_2_wrong value :", *(*cellPtr).ValuePtr)
	}
}

// this test has inputs that should return a value of 1
func TestTwoSampleTTestBuilder_test_1(t *testing.T) {
	err := (*cellPtr).SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetGoodnessPolarity - error message : ", err))
	}
	err = (*cellPtr).SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetMinorThreshold - error message : ", err))
	}
	err = (*cellPtr).SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetMajorThreshold - error message : ", err))
	}
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - SetInputData - error message : ", err))
	}
	var epoch = time.Now().Unix()
	var ctlData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.01,
			Time:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.2,
			Time:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResultPtr *builder.QueryResult = new(builder.QueryResult)
	queryResultPtr.CtlData = &ctlData
	queryResultPtr.ExpData = &expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResultPtr, statistic, "PreCalcRecord")
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}

	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	if *(*cellPtr).ValuePtr != 2 {
		t.Fatal("test_2_wrong value :", *(*cellPtr).ValuePtr)
	}

	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}

	fmt.Println("Pval is", (*cellPtr).Pvalue, "value is ", (*cellPtr).ValuePtr)
	if *((*cellPtr).ValuePtr) != 1 {
		t.Fatal("test_1 wrong value :", (*cellPtr).ValuePtr)
	}
}

// this test has inputs that should return a value of 0
func TestTwoSampleTTestBuilder_test_0(t *testing.T) {
	err := (*cellPtr).SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetGoodnessPolarity - error message : ", err))
	}
	err = (*cellPtr).SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetMinorThreshold - error message : ", err))
	}
	err = (*cellPtr).SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetMajorThreshold - error message : ", err))
	}
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - SetInputData - error message : ", err))
	}

	var epoch = time.Now().Unix()
	var ctlData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.001,
			Time:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.001,
			Time:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResultPtr *builder.QueryResult = new(builder.QueryResult)
	queryResultPtr.CtlData = &ctlData
	queryResultPtr.ExpData = &expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResultPtr, statistic, "PreCalcRecord")
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}

	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - ComputeSignificance - error message : ", err))
	}
	fmt.Println("Pval is", (*cellPtr).Pvalue, "value is ", (*cellPtr).ValuePtr)
	if *((*cellPtr).ValuePtr) != 0 {
		t.Fatal("test_0 wrong value :", (*cellPtr).ValuePtr)
	}
}

func TestTwoSampleTTestBuilder_different_lengths(t *testing.T) {
	err := (*cellPtr).SetGoodnessPolarity(gp)
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetGoodnessPolarity - error message : ", err))
	}
	err = (*cellPtr).SetMinorThreshold(minorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetMinorThreshold - error message : ", err))
	}
	err = (*cellPtr).SetMajorThreshold(majorThreshold)
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetMajorThreshold - error message : ", err))
	}
	if err != nil {
		t.Fatal(fmt.Sprint("SampleTTestBuilder_diff - SetInputData - error message : ", err))
	}

	var epoch = time.Now().Unix()
	var ctlData []interface{}
	for i := 0; i < 10; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.01,
			Time:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData []interface{}
	for i := 0; i < 9; i++ {
		var rec = builder.PreCalcRecord{
			Value: float64(i) * 1.2,
			Time:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResultPtr *builder.QueryResult = new(builder.QueryResult)
	queryResultPtr.CtlData = &ctlData
	queryResultPtr.ExpData = &expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResultPtr, statistic, "PreCalcRecord")
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	err = (*cellPtr).ComputeSignificance()
	if err == nil {
		t.Fatal("TestTwoSampleTTestBuilder_test_identical - ComputeSignificance - no error message : should be 'sample has zero variance'")
	}
}
