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
	"log"
	"github.com/go-playground/validator/v10"
	"github.com/aclements/go-moremath/stats"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

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

// using the Query Result and the statistic perform statistic calculation, perform matching and store the result
func (scc *ScorecardCell) DeriveInputData(qr QueryResult, statisticType string) error {
	//scc.Value = resultPtr
	// use the builder_stats pkg to process the statistic

	// use the builder_stats pkg to perform matching on the input
	return nil
}

func (scc *ScorecardCell) ComputeSignificance() error {
	// scc should hvae already been populated
	if scc.data.CtlPop == nil || scc.data.ExpPop == nil {
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance - no data"))
	}
	// alternate hypothesis is locationDiffers - i.e. null hypothesis is equality.
	var derivedData DerivedDataElement  = scc.data
	alt := stats.LocationDiffers
	// If μ0 is non-zero, this tests if the average of the difference
	// is significantly different from μ0, we assume a zero μ0.
	μ0 := 0.0
	if errs := validate.Var(derivedData.CtlPop, "required"); errs != nil {
		log.Print(errs)
		*scc.valuePtr = -9999  // have to dereference valuePtr - just because
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance", errs))
	}
	if errs := validate.Var(derivedData.ExpPop, "required"); errs != nil {
		log.Print(errs)
		*scc.valuePtr = -9999 // have to dereference valuePtr - just because
		//scc.Value = &v
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
		scc.pvalue = ret.P
		 // have to dereference valuePtr - just because
		*scc.valuePtr, err = deriveValue(scc, difference, ret.P)
		//scc.Value = &v
		if err != nil {
			log.Print(err)
			//scc.Value = &v
			return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance - deriveValue error: ", err))
		}
	} else {
		log.Print(err)
		scc.pvalue = -9999 // pvalue is not a pointer
		*scc.valuePtr = -9999 // have to dereference valuePtr - just because
		//scc.Value = &v
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance", err))
	}
	return nil // no errors
}


// return the internal value that has been derived
func (scc *ScorecardCell) GetValue() int {
	// NOTE: a reference to a non-interface method with a value receiver using
	// a pointer will automatically dereference that pointer
	return *scc.valuePtr
}

func NewTwoSampleTTestBuilder() *ScorecardCell {
	validate = validator.New()
	return &ScorecardCell{}
}
func (scc *ScorecardCell)SetValuePtr(valuePtr int) error {
	if errs := validate.Var(valuePtr, "required"); errs != nil {
		log.Print(errs)
		*scc.valuePtr = -9999
		return errors.New(fmt.Sprint("TwoSampleTTestBuilder ComputeSignificance", errs))
	}
	scc.valuePtr = &valuePtr
	return nil
}

func Build(cellPtr *ScorecardCell, qr QueryResult, statisticType string) error {
	// build the input data elements and
	// for all the input elements fire off a thread to do the compute
	var err error
	err = cellPtr.SetGoodnessPolarity(cellPtr.goodnessPolarity)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director Build SetGoodnessPolarity error ", err))
	}
	err = cellPtr.SetMinorThreshold(cellPtr.minorThreshold)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetminorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(cellPtr.majorThreshold)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetmajorThreshold - error message : ", err))
	}
	err = cellPtr.DeriveInputData(qr, statisticType)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetInputData - error message : ", err))
	}
	 // computes the significance for the data derived in DeriveInputData and stored in cellPtr.data
	err = cellPtr.ComputeSignificance()
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - ComputeSignificance - error message : ", err))
	}
	// insert the elements into the in-memory document
	// upsert the document
	return nil
}

