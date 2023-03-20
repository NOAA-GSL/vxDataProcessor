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

For these analysis we asume for the null hypothesis that the statistic
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

A P-value <= 0.01 (for a 99% major threshold) results in a 2. For 0.01 < P-value <= 0.05
(for a 95% minor threshold) the result is a 1. A P-value greater than the minor threshold
will cause a return of 0.
*/
import (
	"errors"
	"fmt"
	"github.com/aclements/go-moremath/stats"
	"github.com/go-playground/validator/v10"
	"log"
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
func (scc *ScorecardCell) SetGoodnessPolarity(polarity GoodnessPolarity) error {
	errs := validate.Var(polarity, "required,oneof=-1 1")
	if errs != nil {
		log.Print(errs)
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder SetGoodnessPolarity", errs))
	} else {
		scc.goodnessPolarity = polarity
	}
	return nil // no errors
}

// set the major p-value threshold
func (scc *ScorecardCell) SetMajorThreshold(threshold Threshold) error {
	if errs := validate.Var(threshold, "required,gt=0,lt=.5"); errs != nil {
		log.Print(errs)
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder SetMajorThreshold", errs))
	} else {
		scc.majorThreshold = threshold
	}
	return nil // no errors
}

// set the major p-value threshold
func (scc *ScorecardCell) SetMinorThreshold(threshold Threshold) error {
	if errs := validate.Var(threshold, "required,gt=0,lt=.5"); errs != nil {
		log.Print(errs)
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder SetMinorThreshold", errs))
	} else {
		scc.minorThreshold = threshold
	}
	return nil // no errors
}

// get the return value based on the major and minor thresholds compared to the p-value
func deriveValue(scc *ScorecardCell, difference float64, pval float64) (int, error) {
	if errs := validate.Var(difference, "required"); errs != nil {
		log.Print(errs)
		return 0, errors.New(fmt.Sprint("TwoSampleTTestBuilder deriveValue", errs))
	} else {
		if errs := validate.Var(pval, "required"); errs != nil {
			fmt.Println(errs)
			return -9999, errors.New(fmt.Sprint("TwoSampleTTestBuilder deriveValue", errs))
		} else {
			if pval <= float64(scc.majorThreshold) {
				return 2 * int(scc.goodnessPolarity), nil
			}
			if pval <= float64(scc.minorThreshold) {
				return 1 * int(scc.goodnessPolarity), nil
			}
			return 0, nil
		}
	}
}

// using the experimental Query Result and the control QueryResult and the statistic
// perform statistic calculation for each, perform matching and store the resultant  dataSet
func deriveCTCInputData(scc *ScorecardCell, QR *QueryResult, statisticType string) (dataSet DataSet, err error) {
	// derive CTC statistical values for ctl and exp
	var stat float32
	var ctlData []PreCalcRecord
	var expData []PreCalcRecord
	ctlQR := *(QR.CtlData)
	expQR := *(QR.ExpData)
	var record CTCRecord

	for i := 0; i < len(ctlQR); i++ {
		record = (ctlQR)[i].(CTCRecord)
		stat, err = CalculateStatCTC(record.Hit, record.Fa, record.Miss, record.Cn, statisticType)
		if err == nil {
			//include this one
			ctlData = append(ctlData, PreCalcRecord{Value: float64(stat), Time: record.Time})
		} else { /*don't include it*/
		}
	}
	for i := 0; i < len(expQR); i++ {
		record = (expQR)[i].(CTCRecord)
		stat, err = CalculateStatCTC(record.Hit, record.Fa, record.Miss, record.Cn, statisticType)
		if err == nil {
			//include this one
			expData = append(expData, PreCalcRecord{Value: float64(stat), Time: record.Time})
		} else { /*don't include it*/
		}
	}
	// define the dataSet - this is the data struct the holds the two arrays of time and stat value
	dataSet = DataSet{ctlPop: ctlData, expPop: expData}
	// By now we have a dataSet each element of which has only a Time and a Value (i.e. a PreCalcRecord).
	return dataSet, err
}

func deriveScalarInputData(scc *ScorecardCell, qPtr *QueryResult, statisticType string) (dataSet DataSet, err error) {
	// derive Scalar statistical values for ctl and exp
	var stat float64
	var record ScalarRecord
	var ctlData []PreCalcRecord
	var expData []PreCalcRecord
	//var matchedData DataSet
	ctlQR := *(qPtr.CtlData)
	expQR := *(qPtr.ExpData)

	for i := 0; i < len(ctlQR); i++ {
		record = (ctlQR)[i].(ScalarRecord)
		stat, err = CalculateStatScalar(record.SquareDiffSum, record.NSum, record.ObsModelDiffSum, record.ModelSum, record.ObsSum, record.AbsSum, statisticType)
		if err == nil {
			//include this one
			ctlData = append(ctlData, PreCalcRecord{Value: float64(stat), Time: record.Time})
		} else { /*don't include it*/
		}
	}
	for i := 0; i < len(expQR); i++ {
		record = (expQR)[i].(ScalarRecord)
		stat, err = CalculateStatScalar(record.SquareDiffSum, record.NSum, record.ObsModelDiffSum, record.ModelSum, record.ObsSum, record.AbsSum, statisticType)
		if err == nil {
			//include this one
			expData = append(expData, PreCalcRecord{Value: float64(stat), Time: record.Time})
		}
	}
	// return the unmatched Scalar dataSet
	dataSet = DataSet{ctlPop: ctlData, expPop: expData}
	return dataSet, err
}

func derivePreCalcInputData(scc *ScorecardCell, qPtr *QueryResult, statisticType string) (dataSet DataSet, err error) {
	// data is precalculated - don't need to derive stats
	// have to use just the values to create the data set (type DataSet)
	var ctlData PreCalcRecords
	ctlData = make(PreCalcRecords, len(*(qPtr.CtlData)))
	var expData PreCalcRecords
	expData = make(PreCalcRecords, len(*(qPtr.ExpData)))

	for i := 0; i < len(*(qPtr.CtlData)); i++ {
		ctlData = append(ctlData, (*(qPtr.CtlData))[i].(PreCalcRecord))
	}
	for i := 0; i < len(expData); i++ {
		expData = append(expData, (*(qPtr.CtlData))[i].(PreCalcRecord))
	}
	// return the unmatched PreCalculated dataSet
	dataSet = DataSet{ctlPop: ctlData, expPop: expData}
	return dataSet, err
}

func (scc *ScorecardCell) DeriveInputData(qrPtr *QueryResult, statisticType string, dataType string) (err error) {
	var dataSet DataSet
	var matchedDataSet DataSet
	switch dataType {
	case "CTCRecord":
		dataSet, err = deriveCTCInputData(scc, qrPtr, statisticType)
	case "ScalarRecord":
		dataSet, err = deriveScalarInputData(scc, qrPtr, statisticType)
	case "PreCalcRecord":
		dataSet, err = derivePreCalcInputData(scc, qrPtr, statisticType)
	default:
		err = fmt.Errorf("TwoSampleTTestBuilder DeriveInputData unsupported data type: %q", dataType)
	}
	// match the unmatched DataSet
	matchedDataSet, err = GetMatchedDataSet(dataSet)
	// convert matched DataSet to DerivedDataElement
	var de DerivedDataElement
	de.CtlPop = make([]float64, len(matchedDataSet.ctlPop))
	de.ExpPop = make([]float64, len(matchedDataSet.expPop))
	for i := 0; i < len(matchedDataSet.ctlPop); i++ {
		de.CtlPop = append(de.CtlPop, matchedDataSet.ctlPop[i].Value)
		de.ExpPop = append(de.ExpPop, matchedDataSet.expPop[i].Value)
	}
	scc.Data = de
	return err
}

func (scc *ScorecardCell) ComputeSignificance() error {
	// scc should hvae already been populated
	if scc.Data.CtlPop == nil || scc.Data.ExpPop == nil {
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance - no data"))
	}
	// alternate hypothesis is locationDiffers - i.e. null hypothesis is equality.
	var derivedData DerivedDataElement = scc.Data
	alt := stats.LocationDiffers
	// If μ0 is non-zero, this tests if the average of the difference
	// is significantly different from μ0, we assume a zero μ0.
	μ0 := 0.0
	if errs := validate.Var(derivedData.CtlPop, "required"); errs != nil {
		log.Print(errs)
		*scc.ValuePtr = -9999 // have to dereference valuePtr - just because
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance", errs))
	}
	if errs := validate.Var(derivedData.ExpPop, "required"); errs != nil {
		log.Print(errs)
		*scc.ValuePtr = -9999 // have to dereference valuePtr - just because
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance", errs))
	}
	//&TTestResult{N1: n1, N2: n2, T: t, DoF: dof, AltHypothesis: alt, P: p}
	// PairedTTest performs a two-sample paired t-test on samples x1 and x2.
	ret, err := stats.PairedTTest(derivedData.CtlPop, derivedData.ExpPop, μ0, alt)
	if err == nil {
		// what are the means of the populations
		meanCtl := stats.Mean(derivedData.CtlPop)
		meanExp := stats.Mean(derivedData.ExpPop)
		difference := (meanCtl - meanExp)
		scc.Pvalue = ret.P
		// have to dereference valuePtr - just because
		*scc.ValuePtr, err = deriveValue(scc, difference, ret.P)
		//scc.Value = &v
		if err != nil {
			log.Print(err)
			//scc.Value = &v
			return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance - deriveValue error: ", err))
		}
	} else {
		log.Print(err)
		scc.Pvalue = -9999    // pvalue is not a pointer
		*scc.ValuePtr = -9999 // have to dereference valuePtr - just because
		//scc.Value = &v
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance", err))
	}
	return nil // no errors
}

// return the internal value that has been derived
func (scc *ScorecardCell) GetValue() int {
	// NOTE: a reference to a non-interface method with a value receiver using
	// a pointer will automatically dereference that pointer
	return *scc.ValuePtr
}

func NewTwoSampleTTestBuilder() *ScorecardCell {
	validate = validator.New()
	return &ScorecardCell{}
}
func (scc *ScorecardCell) SetValuePtr(valuePtr int) error {
	if errs := validate.Var(valuePtr, "required"); errs != nil {
		log.Print(errs)
		var errorVal int = -9999
		scc.ValuePtr = &errorVal
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance", errs))
	}
	scc.ValuePtr = &valuePtr
	return nil
}

func (scc *ScorecardCell) Build(qrPtr *QueryResult, statisticType string, dataType string) error {
	//DerivePreCalcInputData(ctlQR PreCalcRecords, expQR PreCalcRecords, statisticType string)
	// build the input data elements and
	// for all the input elements fire off a thread to do the compute
	var err error
	err = scc.SetGoodnessPolarity(scc.goodnessPolarity)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director Build SetGoodnessPolarity error ", err))
	}
	err = scc.SetMinorThreshold(scc.minorThreshold)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetminorThreshold - error message : ", err))
	}
	err = scc.SetMajorThreshold(scc.majorThreshold)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetmajorThreshold - error message : ", err))
	}
	err = scc.DeriveInputData(qrPtr, statisticType, dataType)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetInputData - error message : ", err))
	}
	// computes the significance for the data derived in DeriveInputData and stored in cellPtr.data
	err = scc.ComputeSignificance()
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - ComputeSignificance - error message : ", err))
	}
	// insert the elements into the in-memory document
	// upsert the document
	return nil
}
