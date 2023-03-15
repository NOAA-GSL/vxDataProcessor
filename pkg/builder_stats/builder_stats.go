package builder_stats

import (
    "fmt"
    "math"
    "github.com/go-playground/validator/v10"
)
// if fields contains square_diff_sum => QueryResult is ScalarRecord
// else if fields contain hits => QueryResult is CTCRecord
// else QueryResult is PreCalcRecord
type CTCRecord = struct {
	hit int
	miss int
	fa int
	cn int
	time int64
}
type ScalarRecord = struct {
	squareDiffSum float64
	NSum float64
	obsModelDiffSum float64
	modelSum float64
	obsSum float64
	absSum float64
	time float64
}
type PreCalcRecord struct {
	value float64
	time int64
}

type DataSet struct{
    ctlPop []PreCalcRecord
    expPop []PreCalcRecord
}

var validate *validator.Validate

/*
These are stats functions that are used to derive scorecard stats from raw populations.
There is also a time matching function. These functions are used by the builder functions.
*/

// calculates the statistic for ctc plots
func calculateStatCTC(hit float32, fa float32, miss float32, cn float32, statistic string) (float64, error){
    if errs := validate.Var(hit, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(fa, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(cn, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(miss, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(statistic, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    var value float64
    switch (statistic) {
        case "TSS (True Skill Score)": //radar
            value = ((hit * cn - fa * miss) / ((hit + miss) * (fa + cn))) * 100;
        // some PODy measures look for a value over a threshold, some look for under
        case "PODy (POD of value < threshold)": //ceiling
        case "PODy (POD of value > threshold)": //radar
            value = hit / (hit + miss) * 100;
        // some PODn measures look for a value under a threshold, some look for over
        case "PODn (POD of value > threshold)": //ceiling
        case "PODn (POD of value < threshold)": // radar
            value = cn / (cn + fa) * 100;
        case "FAR (False Alarm Ratio)": // radar
            value = fa / (fa + hit) * 100;
        case "CSI (Critical Success Index)": // radar
            value = hit / (hit + miss + fa) * 100;
        case "HSS (Heidke Skill Score)": // radar
            value = 2 * (cn * hit - miss * fa) / ((cn + fa) * (fa + hit) + (cn + miss) * (miss + hit)) * 100;
        case "ETS (Equitable Threat Score)": // radar
            value = (hit - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) / ((hit + fa + miss) - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) * 100;
        default:
            return 0, fmt.Errorf("builder_stats.calculateStatCTC: %q %q", "Invalid statistic:", statistic)
    }
    return value, nil;
}

// calculates the statistic for scalar partial sums plots
func calculateStatScalar (squareDiffSum, NSum, obsModelDiffSum, modelSum, obsSum, absSum float64, statistic string)(float64, error) {
    if errs := validate.Var(squareDiffSum, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(NSum, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(obsModelDiffSum, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(modelSum, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(obsSum, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(absSum, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    if errs := validate.Var(statistic, "required"); errs != nil {
		return 0, fmt.Errorf("builder_stats calculateStatCTC %q", errs)
    }
    var value float64
    switch (statistic) {
        case "RMSE": //surface
            value = math.Sqrt(squareDiffSum / NSum);
            break;
        case "Bias (Model - Obs)": //surface
            value = (modelSum - obsSum) / NSum;
            break;
        case "MAE (temp and dewpoint only)": //surface
        case "MAE": // landuse
            value = absSum / NSum;
            break;
    }
    return value, nil;

}

// function for removing unmatched data from a dataset containing two curves
// The intersection of the ctlData and the expData based on the time elements.
// This function assumes that the two slices are sorted by the time element (which is an epoch)
func GetMatchedDataSet(data DataSet)(DataSet, error){
    var result DataSet
    var indexCtl int = 0
    var indexExp int = 0
    var maxLen = int(math.Max(float64(len(data.ctlPop)),float64(len(data.expPop))))
    var resultIndex int = 0
     for {
        if data.ctlPop[indexCtl].time == data.expPop[indexExp].time {
            // time matches and valid values so append to result
            result.ctlPop[resultIndex].time=data.ctlPop[indexCtl].time
            result.ctlPop[resultIndex].value=data.ctlPop[indexCtl].value
            result.expPop[resultIndex].time=data.expPop[indexCtl].time
            result.expPop[resultIndex].value=data.expPop[indexCtl].value
            // increment indexes
            indexCtl++
            indexExp++
        } else {
            // times did not match - increment the earliest one
            if data.ctlPop[indexCtl].time < data.expPop[indexExp].time {
                // increment the ctlPop index
                indexCtl++
            } else {
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
    return result, nil
}

