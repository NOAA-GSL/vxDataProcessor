package builder

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"math"
)

// these are floats because of the division in the CalculateStatCTC func
type CTCRecord struct {
	Hit  float32
	Miss  float32
	Fa   float32
	Cn   float32
	Time int64
}
type CTCRecords = []CTCRecord

type ScalarRecord struct {
	SquareDiffSum   float64
	NSum            float64
	ObsModelDiffSum float64
	ModelSum        float64
	ObsSum          float64
	AbsSum          float64
	Time            int64
}
type ScalarRecords []ScalarRecord

type PreCalcRecord struct {
	Value float64
	Time  int64
}
type PreCalcRecords []PreCalcRecord

type DataSet struct {
	ctlPop []PreCalcRecord
	expPop []PreCalcRecord
}

var validate *validator.Validate

/*
These are stats functions that are used to derive scorecard stats from raw populations.
There is also a time matching function. These functions are used by the builder functions.
*/

// calculates the statistic for ctc plots
func CalculateStatCTC(hit float32, fa float32, miss float32, cn float32, statistic string) (float32, error) {
	var err error
	var value float32
  validate = validator.New()
  if err = validate.Var(hit, "gte=0"); err != nil {
    value = 0
		return value, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(fa, "gte=0"); err != nil {
    value = 0
		return value, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(cn, "gte=0"); err != nil {
    value = 0
		return value, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(miss, "gte=0"); err != nil {
    value = 0
		return value, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(statistic, "gte=0"); err != nil {
    value = 0
    return value, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}

  switch statistic {
	case "TSS (True Skill Score)": //radar
		value = ((hit*cn - fa*miss) / ((hit + miss) * (fa + cn))) * 100
	// some PODy measures look for a value over a threshold, some look for under
	case "PODy (POD of value < threshold)": //ceiling
    value = hit / (hit + miss) * 100
	case "PODy (POD of value > threshold)": //radar
		value = hit / (hit + miss) * 100
	// some PODn measures look for a value under a threshold, some look for over
	case "PODn (POD of value > threshold)": //ceiling
    value = cn / (cn + fa) * 100
	case "PODn (POD of value < threshold)": // radar
		value = cn / (cn + fa) * 100
	case "FAR (False Alarm Ratio)": // radar
		value = fa / (fa + hit) * 100
	case "CSI (Critical Success Index)": // radar
		value = hit / (hit + miss + fa) * 100
	case "HSS (Heidke Skill Score)": // radar
		value = 2 * (cn*hit - miss*fa) / ((cn+fa)*(fa+hit) + (cn+miss)*(miss+hit)) * 100
	case "ETS (Equitable Threat Score)": // radar
		value = (hit - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) / ((hit + fa + miss) - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) * 100
	default:
		err = fmt.Errorf(fmt.Sprintf("builder_stats.calculateStatCTC: %q %q", "Invalid statistic:", statistic))
    return 0, err
	}
  if math.IsNaN(float64(value)){
    err = fmt.Errorf("builder_stats.calculateStatCTC value is NaN")
  }
  if math.IsInf(float64(value),0) {
    err = fmt.Errorf("builder_stats.calculateStatCTC value is Infinity")
  }
  return value, err
}

// calculates the statistic for scalar partial sums plots
func CalculateStatScalar(squareDiffSum, NSum, obsModelDiffSum, modelSum, obsSum, absSum float64, statistic string) (float64, error) {
	var err error
	var value float64
	if err = validate.Var(squareDiffSum, "required"); err != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(NSum, "required"); err != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(obsModelDiffSum, "required"); err != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(modelSum, "required"); err != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(obsSum, "required"); err != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(absSum, "required"); err != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	if err = validate.Var(statistic, "required"); err != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", err)
	}
	switch statistic {
	case "RMSE": //surface
		value = math.Sqrt(squareDiffSum / NSum)
		break
	case "Bias (Model - Obs)": //surface
		value = (modelSum - obsSum) / NSum
		break
	case "MAE (temp and dewpoint only)": //surface
	case "MAE": // landuse
		value = absSum / NSum
		break
	}
	return value, err

}

// function for removing unmatched data from a dataset containing two curves
// The intersection of the ctlData and the expData based on the time elements.
// This function assumes that the two slices are sorted by the time element (which is an epoch)
// The DataSet consists of time and value elements only, since the statistical value has
// already been derived
func GetMatchedDataSet(dataSet DataSet) (DataSet, error) {
	var result DataSet
	var indexCtl int = 0
	var indexExp int = 0
	var maxLen = int(math.Max(float64(len(dataSet.ctlPop)), float64(len(dataSet.expPop))))
	var resultIndex int = 0
	var err error = nil
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("builder_stat calculateStatCTC recovered panic - probably divide by 0")
		}
	}()

		for {
			if dataSet.ctlPop[indexCtl].Time == dataSet.expPop[indexExp].Time {
				// time matches and valid values so append to result
				result.ctlPop[resultIndex] = dataSet.ctlPop[indexCtl]
				result.expPop[resultIndex] = dataSet.expPop[indexExp]
				indexCtl++
				indexExp++
			} else {
				// times did not match - increment the earliest one
				if result.ctlPop[indexCtl].Time < result.expPop[indexExp].Time {
					// increment the ctlPop index
					indexCtl++
				} else {
					// increment the expPop index
					indexExp++
				}
				// continue with new index
				continue
			}
			resultIndex++
			if resultIndex >= maxLen {
				break
			}
		}
	return result, err
}