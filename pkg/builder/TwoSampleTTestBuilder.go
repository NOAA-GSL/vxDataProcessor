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
https://online.stat.psu.edu/statprogram/reviews/statistical-concepts/hypothesis-testing/p-value-approach

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
	"fmt"
	"github.com/aclements/go-moremath/stats"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate = validator.New()

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
func (br *builderResult)setGoodnessPolarity(polarity GoodnessPolarity) Builder {
	errs := validate.Var(polarity, "required,oneof=-1 1")
	if errs != nil {
		fmt.Println(errs)
	} else {
		br.goodnessPolarity = polarity
	}
	return nil
}

// set the major p-value threshold
func (br *builderResult)setMajorThreshold(threshold Threshold) Builder  {
	if errs := validate.Var(threshold, "required,gt=0,lt=.5"); errs != nil {
		fmt.Println(errs)
	} else {
		br.majorThreshold = threshold
	}
	return nil
}

// set the major p-value threshold
func (br *builderResult)setMinorThreshold(threshold Threshold)  Builder {
	if errs := validate.Var(threshold, "required,gt=0,lt=.5"); errs != nil {
		fmt.Println(errs)
	} else {
		br.minorThreshold = threshold
	}
	return nil
}

// get the return value based on the major and minor thresholds compared to the p-value
func getValue(br builderResult, difference float64, pval float64) int {
	if errs := validate.Var(difference, "required"); errs != nil {
		fmt.Println(errs)
		return 0 // ??? don't know what to return for errors
	} else {
		if errs := validate.Var(pval, "required"); errs != nil {
			fmt.Println(errs)
			return 0 // ??? don't know what to return for errors
		} else {
			if pval <= float64(br.majorThreshold) {
				return 2 * int(br.goodnessPolarity)
			}
			if pval <= float64(br.minorThreshold) {
				return 1 * int(br.goodnessPolarity)
			}
			return 0
		}
	}
}

func (br *builderResult)computeSignificance(derivedData DerivedData) Builder {
	// alternate hypothesis is locationDiffers - i.e. null hypothesis is equality.
	alt := stats.LocationDiffers
	// If μ0 is non-zero, this tests if the average of the difference
	// is significantly different from μ0, we assume a zero μ0.
	μ0 := 0.0
	if errs := validate.Var(derivedData.CtlPop, "required,len gt 0"); errs != nil {
		fmt.Println(errs)
		return nil // ??? don't know what to return for errors
	}
	if errs := validate.Var(derivedData.ExpPop, "required,len gt 0"); errs != nil {
		fmt.Println(errs)
		return nil // ??? don't know what to return for errors
	}
	//&TTestResult{N1: n1, N2: n2, T: t, DoF: dof, AltHypothesis: alt, P: p}
	// PairedTTest performs a two-sample paired t-test on samples x1 and x2.
	ret, err := stats.PairedTTest(derivedData.CtlPop, derivedData.ExpPop, μ0, alt)
	if err != nil {
		// what are the means of the populations
		meanCtl := stats.Mean(derivedData.CtlPop)
		meanExp := stats.Mean(derivedData.ExpPop)
		difference := (meanCtl - meanExp)
		br.value = getValue(*br, difference, ret.P)
		return nil
	} else {
		fmt.Println(err)
		return nil
	}
}

func (br *builderResult)deriveData(inputData InputData) Builder {
	// put the code to derive the data from the inputData HERE!
	data := DerivedData{
		// caclulate data here
		CtlPop: nil,
		ExpPop: nil,
	}
	br.data = data
	return nil
}

func (br *builderResult) Build() Builder {
	return &builderResult{
		data: br.data,
		goodnessPolarity: br.goodnessPolarity,
		majorThreshold: br.majorThreshold,
		minorThreshold: br.minorThreshold,
		value: br.value,
	}
}

func NewTwoSampleTTestBuilder() Builder {
	return &builderResult{}
}