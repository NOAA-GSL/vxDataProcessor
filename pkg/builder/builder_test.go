package builder

import (
	"fmt"
	"testing"
	"time"
)

var gp = GoodnessPolarity(1)
var minorThreshold = Threshold(0.05)
var majorThreshold = Threshold(0.01)
var cellPtr = GetBuilder("TwoSampleTTest")

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
	var ctlData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(i) * 1.1,
			avtime:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(i) * 1.1,
			avtime:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResult BuilderPreCalcResult
	queryResult.CtlData = ctlData
	queryResult.ExpData = expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResult, statistic)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}

	// expect an error here - sample has zero variance
	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal("TestTwoSampleTTestBuilder_test_identical - ComputeSignificance - error message : ", err)
	}
}

// this test has inputs that should return a value of 2
func TestTwoSampleTTestBuilder_test_2(t *testing.T) {
	var cellPtr = NewTwoSampleTTestBuilder()
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
	var ctlData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(i) * 1.01,
			avtime:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(i) * 1.2,
			avtime:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResult BuilderPreCalcResult
	queryResult.CtlData = ctlData
	queryResult.ExpData = expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResult, statistic)
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
	var normData = [10]int{86, 74, 79, 94, 73, 92, 66, 77, 74, 78}
	var ctlData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(normData[i]),
			avtime:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	// I don't claim to know why but this modification gives a number set that generates a pvalue 0.015523870374046123 which results in value 1
	var expData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(normData[i] * (i % 2)),
			avtime:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResult BuilderPreCalcResult
	queryResult.CtlData = ctlData
	queryResult.ExpData = expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResult, statistic)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	fmt.Println("Pval is", (*cellPtr).Pvalue, "value is ", *(*cellPtr).ValuePtr)
	if *((*cellPtr).ValuePtr) != 1 {
		t.Fatal("test_1 wrong value :", *(*cellPtr).ValuePtr)
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
	var normData = [10]int{86, 74, 79, 94, 73, 92, 66, 77, 74, 78}
	var ctlData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(normData[i]),
			avtime:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	// The first element is off by one
	var expData PreCalcRecords
	normData = [10]int{87, 74, 79, 94, 73, 92, 66, 77, 74, 78}
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(normData[i]),
			avtime:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResult BuilderPreCalcResult
	queryResult.CtlData = ctlData
	queryResult.ExpData = expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResult, statistic)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - ComputeSignificance - error message : ", err))
	}
	fmt.Println("Pval is", (*cellPtr).Pvalue, "value is ", *(*cellPtr).ValuePtr)
	if *((*cellPtr).ValuePtr) != 0 {
		t.Fatal("test_0 wrong value :", *(*cellPtr).ValuePtr)
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
	var ctlData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(i) * 1.01,
			avtime:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	var expData PreCalcRecords
	for i := 0; i < 9; i++ {
		var rec = PreCalcRecord{
			stat: float64(i) * 1.2,
			avtime:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResult BuilderPreCalcResult
	queryResult.CtlData = ctlData
	queryResult.ExpData = expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResult, statistic)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_0 - ComputeSignificance - error message : ", err))
	}
}

// this test has inputs that should return a value of 1 after matching (ctl missing one element)
func TestTwoSampleTTestBuilder_test__match_ctl_short_1(t *testing.T) {
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
	var normData = [10]int{86, 74, 79, 94, 73, 92, 66, 77, 74, 78}
	var ctlData PreCalcRecords
	for i := 0; i < 10; i++ {
		// skip number 5
		if i == 5 {
			continue
		}
		var rec = PreCalcRecord{
			stat: float64(normData[i]),
			avtime:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	// I don't claim to know why but this modification gives a number set that generates a pvalue 0.015523870374046123 which results in value 1
	var expData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(normData[i] * (i % 2)),
			avtime:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResult BuilderPreCalcResult
	queryResult.CtlData = ctlData
	queryResult.ExpData = expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResult, statistic)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	fmt.Println("Pval is", (*cellPtr).Pvalue, "value is ", *(*cellPtr).ValuePtr)
	if *((*cellPtr).ValuePtr) != 1 {
		t.Fatal("test_1 wrong value :", *(*cellPtr).ValuePtr)
	}
}

// this test has inputs that should return a value of 1 after matching (exp missing one element)
func TestTwoSampleTTestBuilder_test__match_exp_short_1(t *testing.T) {
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
	var normData = [10]int{86, 74, 79, 94, 73, 92, 66, 77, 74, 78}
	var ctlData PreCalcRecords
	for i := 0; i < 10; i++ {
		var rec = PreCalcRecord{
			stat: float64(normData[i]),
			avtime:  int64(i) + epoch,
		}
		ctlData = append(ctlData, rec)
	}
	// I don't claim to know why but this modification gives a number set that generates a pvalue 0.015523870374046123 which results in value 1
	var expData PreCalcRecords
	for i := 0; i < 10; i++ {
		// skip number 5
		if i == 5 {
			continue
		}
		var rec = PreCalcRecord{
			stat: float64(normData[i] * (i % 2)),
			avtime:  int64(i) + epoch,
		}
		expData = append(expData, rec)
	}
	var queryResult BuilderPreCalcResult
	queryResult.CtlData = ctlData
	queryResult.ExpData = expData
	var statistic string = "TSS (True Skill Score)"
	err = (*cellPtr).DeriveInputData(queryResult, statistic)
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	err = (*cellPtr).ComputeSignificance()
	if err != nil {
		t.Fatal(fmt.Sprint("TestTwoSampleTTestBuilder_test_1 - ComputeSignificance - error message : ", err))
	}
	fmt.Println("Pval is", (*cellPtr).Pvalue, "value is ", *(*cellPtr).ValuePtr)
	if *((*cellPtr).ValuePtr) != 1 {
		t.Fatal("test_1 wrong value :", *(*cellPtr).ValuePtr)
	}
}
