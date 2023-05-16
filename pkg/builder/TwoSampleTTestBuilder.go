package builder

/* This is a PairedTTest builder.
The PairedTTest returns a TTestResult,
TTestResult{N1: n1, N2: n2, T: t, DoF: dof, AltHypothesis: alt, P: p}
where N1 and N2 are the size of the populations A and B respectively,
T is the actual value of the t-statistic, dof(degree of freedom) is 0,
and P is the p-value. The AltHypothesis (for us) is always set to
LocationDiffers which specifies the alternative hypothesis that
the locations of the two samples are not equal. This is a
two-tailed test.
For reference about Hypothesis testing with P-value look here...
https://refactoring.guru/design-patterns/builder/go/example

For these analysis we assume for the null hypothesis that the statistic
that is generated from the "validation data source", which might be thought of as the
control source population, is the same as the "data source", which might be thought
of as the experimental source population.
This builder's goal is to try to demonstrate a likely difference and assign a
number between -2 and 2 for the P-value result. The positive or negative indicator depends
on the statistic being examined. A positive indicator is considered "good" and a negative
indicator is "bad". To determine the sign we take the mean of the experimental
"data source" and subtract the mean of the control "validation data source".
For a statistic like RMSE or BIAS a positive difference is "bad" and a negative difference
is "good" because we want to minimize the error or the bias in the experiment.
For CSI "Critical Success Index" it would be the opposite because CSI ranges from 0
which is poor to 1 which is good. A return value of 0 is neutral / insignificant.

For p-value indicators: the smaller the p-value the more likely the null hypothesis, so
the stronger is the evidence against the null hypothesis, therefore the larger the indicator.
Therefore a small indicator represents weaker evidence against the null hypothesis and
a large indicator represents stronger evidence against the null hypothesis.
A P-value <= 0.01 (for a 99% major threshold) results in a 2. For 0.01 < P-value <= 0.05
(for a 95% minor threshold) the result is a 1. A P-value greater than the minor threshold
will cause a return of 0.
*/
import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strings"
	"sync"

	"github.com/aclements/go-moremath/stats"
	"github.com/go-playground/validator/v10"
)

// use a single instance of Validate, it caches struct info

// setters:
// The goodnessPolarity indicates if this population is positive good (like for TS/CSI)
// or negative good like for RMSE or BIAS. The null hypothesis is that the populations
// are identical so a positive difference for an error value is bad
// and a negative difference is good (less error in the experimental population
// is good for error). If the statistic is Threat Score / Critical
// Succes Index then a positive difference is good (0 is the worst 1 is the best)
// and a negative difference in the experimental population would be bad.
// This information is determined outside of this builder (the builder doesn't know
// what parameter combination is being tested) so the builder must be told
// what the "goodnes polarity" is +1 or -1.

func (scc *ScorecardCell) setGoodnessPolarity(polarity GoodnessPolarity) error {
	errs := validate.Var(polarity, "required,oneof=-1 1")
	if errs != nil {
		log.Print(errs)
		return fmt.Errorf("TwoSampleTTestBuilder SetGoodnessPolarity %w", errs)
	}
	scc.goodnessPolarity = polarity
	return nil // no errors
}

// set the major p-value threshold
func (scc *ScorecardCell) setMajorThreshold(threshold Threshold) error {
	scc.majorThreshold = threshold
	return nil // no errors
}

// set the major p-value threshold
func (scc *ScorecardCell) setMinorThreshold(threshold Threshold) error {
	scc.minorThreshold = threshold
	return nil // no errors
}

// set the keychain
func (scc *ScorecardCell) SetKeyChain(keychain []string) error {
	scc.keychain = keychain
	return nil // no errors
}

// set the statisticType
func (scc *ScorecardCell) SetStatisticType(statisticType StatisticType) error {
	scc.statisticType = statisticType
	return nil // no errors
}

// get the return value based on the major and minor thresholds compared to the p-value.
// If the difference is negative and the goodnessPolarity is positive then the result
// value is negative. If the difference is positive and the goodnessPolarity is negative
// then the result value is negative. If the difference is positive and the goodnessPolarity
// is positive then the result value is positive. If the difference is negative and the
// goodnessPolarity is negative then the result value is positive.

func (scc *ScorecardCell) deriveValue(difference float64, pval float64) (int, error) {
	diffSign := 1
	if difference < 0 {
		diffSign = -1
	}
	if pval <= float64(100-scc.majorThreshold) {
		return 2 * diffSign * int(scc.goodnessPolarity), nil
	}
	if pval <= float64(100-scc.minorThreshold) {
		return 1 * diffSign * int(scc.goodnessPolarity), nil
	}
	return 0, nil
}

// set the value field - controlled by mutex
func (scc *ScorecardCell) setValue(value int) {
	scc.value = value
}

// using the experimental Query Result and the control QueryResult and the statistic
// perform statistic calculation for each, perform matching and store the resultant  dataSet
func (scc *ScorecardCell) deriveCTCInputData(queryResult BuilderCTCResult, statisticType StatisticType) (dataSet DataSet, err error) {
	// derive CTC statistical values for ctl and exp
	var stat float32
	var ctlData PreCalcRecords
	var expData PreCalcRecords
	var record CTCRecord

	for i := 0; i < len(queryResult.CtlData); i++ {
		record = queryResult.CtlData[i]
		stat, err = calculateStatCTC(record.Hit, record.Fa, record.Miss, record.Cn, statisticType)
		if err == nil {
			// include this one
			ctlData = append(ctlData, PreCalcRecord{Stat: float64(stat), Avtime: record.Avtime})
		}
	}
	for i := 0; i < len(queryResult.ExpData); i++ {
		record = queryResult.ExpData[i]
		stat, err = calculateStatCTC(record.Hit, record.Fa, record.Miss, record.Cn, statisticType)
		if err == nil {
			// include this one
			expData = append(expData, PreCalcRecord{Stat: float64(stat), Avtime: record.Avtime})
		}
	}
	// define the dataSet - this is the data struct the holds the two arrays of time and stat value
	dataSet = DataSet{ctlPop: ctlData, expPop: expData}
	// By now we have a dataSet each element of which has only a Time and a Value (i.e. a PreCalcRecord).
	return dataSet, err
}

func (scc *ScorecardCell) deriveScalarInputData(queryResult BuilderScalarResult, statisticType StatisticType) (dataSet DataSet, err error) {
	// derive Scalar statistical values for ctl and exp
	var stat float64
	var ctlData []PreCalcRecord
	var expData []PreCalcRecord
	var record ScalarRecord

	for _, record = range queryResult.CtlData {
		stat, err = calculateStatScalar(record.SquareDiffSum, record.NSum, record.ObsModelDiffSum, record.ModelSum, record.ObsSum, record.AbsSum, statisticType)
		if err == nil {
			// include this one
			ctlData = append(ctlData, PreCalcRecord{Stat: stat, Avtime: record.Avtime})
		}
	}
	for _, record = range queryResult.ExpData {
		stat, err = calculateStatScalar(record.SquareDiffSum, record.NSum, record.ObsModelDiffSum, record.ModelSum, record.ObsSum, record.AbsSum, statisticType)
		if err == nil {
			// include this one
			expData = append(expData, PreCalcRecord{Stat: stat, Avtime: record.Avtime})
		}
	}
	// return the unmatched Scalar dataSet
	dataSet = DataSet{ctlPop: ctlData, expPop: expData}
	return dataSet, err
}

func (scc *ScorecardCell) derivePreCalcInputData(queryResult BuilderPreCalcResult, statisticType StatisticType) (dataSet DataSet, err error) {
	// data is precalculated - don't need to derive stats
	// have to use just the values to create the data set (type DataSet)
	var ctlData PreCalcRecords
	var expData PreCalcRecords
	// set the statistic type for future reference
	scc.statisticType = statisticType
	ctlData = append(ctlData, queryResult.CtlData...)
	expData = append(expData, queryResult.ExpData...)
	// return the unmatched PreCalculated dataSet
	dataSet = DataSet{ctlPop: ctlData, expPop: expData}
	return dataSet, err
}

func (scc *ScorecardCell) deriveInputData(qrPtr interface{}) (err error) {
	var dataSet DataSet
	var matchedDataSet DataSet
	dataType := reflect.TypeOf(qrPtr).Name()
	switch dataType {
	case "BuilderCTCResult":
		dataSet, err = scc.deriveCTCInputData(qrPtr.(BuilderCTCResult), scc.statisticType)
	case "BuilderScalarResult":
		dataSet, err = scc.deriveScalarInputData(qrPtr.(BuilderScalarResult), scc.statisticType)
	case "BuilderPreCalcResult":
		dataSet, err = scc.derivePreCalcInputData(qrPtr.(BuilderPreCalcResult), scc.statisticType)
	default:
		err = fmt.Errorf("TwoSampleTTestBuilder DeriveInputData unsupported data type: %q", dataType)
	}
	if err != nil {
		return err
	}
	// match the unmatched DataSet
	matchedDataSet, err = getMatchedDataSet(dataSet)
	// convert matched DataSet to DerivedDataElement
	var de DerivedDataElement
	for i := 0; i < len(matchedDataSet.ctlPop); i++ {
		de.CtlPop = append(de.CtlPop, matchedDataSet.ctlPop[i].Stat)
		de.ExpPop = append(de.ExpPop, matchedDataSet.expPop[i].Stat)
	}
	scc.Data = de
	return err
}

func (scc *ScorecardCell) computeSignificance() error {
	// scc should have already been populated
	if scc.Data.CtlPop == nil || scc.Data.ExpPop == nil {
		// if there is no data then set the value to fillValue and move on
		scc.setValue(ErrorValue)
		return nil // no errors
	}
	// alternate hypothesis is locationDiffers - i.e. null hypothesis is equality.
	var derivedData DerivedDataElement = scc.Data
	alt := stats.LocationDiffers
	// If μ0 is non-zero, this tests if the average of the difference
	// is significantly different from μ0, we assume a zero μ0.
	μ0 := 0.0
	if errs := validate.Var(derivedData.CtlPop, "required"); errs != nil {
		log.Print(errs)
		var v int = ErrorValue
		scc.setValue(v)
		return fmt.Errorf("TwoSampleTTestBuilder ComputeSignificance %w", errs)
	}
	if errs := validate.Var(derivedData.ExpPop, "required"); errs != nil {
		log.Print(errs)
		var v int = ErrorValue
		scc.setValue(v)
		return fmt.Errorf("TwoSampleTTestBuilder ComputeSignificance %w", errs)
	}
	//&TTestResult{N1: n1, N2: n2, T: t, DoF: dof, AltHypothesis: alt, P: p}
	// PairedTTest performs a two-sample paired t-test on samples x1 and x2.
	ret, err := stats.PairedTTest(derivedData.CtlPop, derivedData.ExpPop, μ0, alt)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "zero variance") {
			// we are not considering identical sets to be errors
			// set pval to 1 and value to 0
			scc.pvalue = 1
			var v int = 0
			scc.setValue(v)
			return nil
		} else {
			log.Print(err)
			scc.pvalue = ErrorValue
			var v int = ErrorValue
			scc.setValue(v)
			return fmt.Errorf("TwoSampleTTestBuilder ComputeSignificance %w", err)
		}
	} else {
		// what are the means of the populations?
		meanCtl := stats.Mean(derivedData.CtlPop)
		meanExp := stats.Mean(derivedData.ExpPop)
		var difference float64
		if scc.statisticType == Bias_Model_Obs {
			difference = (math.Abs(meanCtl) - math.Abs(meanExp))
		} else {
			difference = (meanExp - meanCtl)
		}
		scc.pvalue = ret.P
		v, err := scc.deriveValue(difference, ret.P)
		if err != nil {
			log.Print(err)
			return fmt.Errorf("TwoSampleTTestBuilder ComputeSignificance - deriveValue error:  %w", err)
		}
		scc.setValue(v)
	}
	return nil // no errors
}

// return the internal value that has been derived
func (scc *ScorecardCell) getValue() int {
	// NOTE: a reference to a non-interface method with a value receiver using
	// a pointer will automatically dereference that pointer
	return scc.value
}

func NewTwoSampleTTestBuilder() *ScorecardCell {
	validate = validator.New()
	return &ScorecardCell{mu: sync.Mutex{}}
}

func getGoodnessPolarity(statisticType StatisticType) (polarity GoodnessPolarity, err error) {
	/*
		    see builder/iBuilder.go for the definitions of StatisticType and GoodnessPolarity
			"RMSE": "Want control to exceed experimental" 1
			"Bias (Model - Obs)": "Want control to exceed experimental" 1
			"MAE (temp and dewpoint only)": "Want control to exceed experimental" 1
			"MAE": "Want control to exceed experimental" 1
			"TSS (True Skill Score)": "Want experimental to exceed control" -1
			"PODy (POD of value < threshold)": "Want experimental to exceed control" -1
			"PODy (POD of value>threshold)": "Want experimental to exceed control" -1
			"PODn (POD of value>threshold)": "Want experimental to exceed control" -1
			"PODn (POD of value < threshold)": "Want experimental to exceed control" -1
			"FAR (False Alarm Ratio)": "Want control to exceed experimental" 1
			"CSI (Critical Success Index)": "Want experimental to exceed control" -1
			"HSS (Heidke Skill Score)": "Want experimental to exceed control" -1
			"ETS (Equitable Threat Score)": "Want experimental to exceed control" -1
			"ACC": "Want experimental to exceed control" -1
	*/

	switch statisticType {
	case RMSE:
		return 1, nil
	case Bias_Model_Obs:
		return 1, nil
	case MAE:
		return 1, nil
	case MAE_temp_and_dewpoint_only:
		return 1, nil
	case TSS_True_Skill_Score:
		return -1, nil
	case PODy_POD_of_value_lt_threshold:
		return -1, nil
	case PODy_POD_of_value_gt_threshold:
		return -1, nil
	case PODn_POD_of_value_gt_threshold:
		return -1, nil
	case PODn_POD_of_value_lt_threshold:
		return -1, nil
	case FAR_False_Alarm_Ratio:
		return 1, nil
	case CSI_Critical_Success_Index:
		return -1, nil
	case HSS_Heidke_Skill_Score:
		return -1, nil
	case ETS_Equitable_Threat_Score:
		return -1, nil
	case ACC:
		return -1, nil
	default:
		return -1, fmt.Errorf("TwoSampleTTestBuilder getGoodnessPolarity unknown statistic %q", statisticType)
	}
}

func (scc *ScorecardCell) Build(qrPtr interface{}, statisticType StatisticType, minorThreshold float64, majorThreshold float64) (value int, err error) {
	// DerivePreCalcInputData(ctlQR PreCalcRecords, expQR PreCalcRecords, statisticType string)
	// build the input data elements and

	scc.statisticType = statisticType
	goodnessPolarity, err := getGoodnessPolarity(statisticType)
	if err != nil {
		return ErrorValue, fmt.Errorf("mysql_director Build SetGoodnessPolarity error  %w", err)
	}
	err = scc.setGoodnessPolarity(goodnessPolarity)
	if err != nil {
		return ErrorValue, fmt.Errorf("mysql_director Build SetGoodnessPolarity error  %w", err)
	}

	err = scc.setMinorThreshold(Threshold(minorThreshold))
	if err != nil {
		return ErrorValue, fmt.Errorf("mysql_director Build SetMinorThreshold error  %w", err)
	}

	err = scc.setMajorThreshold(Threshold(majorThreshold))
	if err != nil {
		return ErrorValue, fmt.Errorf("mysql_director Build SetMajorThreshold error  %w", err)
	}

	err = scc.deriveInputData(qrPtr)
	if err != nil {
		return ErrorValue, fmt.Errorf("mysql_director - build - SetInputData - error message :  %w", err)
	}
	// computes the significance for the data derived in DeriveInputData and stored in cellPtr.data
	err = scc.computeSignificance()
	if err != nil {
		return ErrorValue, fmt.Errorf("mysql_director - build - ComputeSignificance - error message :  %w", err)
	}
	// insert the elements into the result
	return scc.getValue(), nil
	// manager will upsert the document
}
