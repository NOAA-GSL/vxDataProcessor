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
func calculateStatCTC(hit int, fa int, miss int, cn int, statistic string) (int, error){
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
    var value int
    switch (statistic) {
        case "TSS (True Skill Score)":
            value = ((hit * cn - fa * miss) / ((hit + miss) * (fa + cn))) * 100;
        // some PODy measures look for a value over a threshold, some look for under
        case "PODy (POD of value < threshold)":
        case "PODy (POD of value > threshold)":
            value = hit / (hit + miss) * 100;
        // some PODn measures look for a value under a threshold, some look for over
        case "PODn (POD of value > threshold)":
        case "PODn (POD of value < threshold)":
            value = cn / (cn + fa) * 100;
        case "POFD (Probability of False Detection)":
            value = fa / (fa + cn) * 100;
        case "FAR (False Alarm Ratio)":
            value = fa / (fa + hit) * 100;
        case "Bias (forecast/actual)":
            value = (hit + fa) / (hit + miss);
        case "CSI (Critical Success Index)":
            value = hit / (hit + miss + fa) * 100;
        case "HSS (Heidke Skill Score)":
            value = 2 * (cn * hit - miss * fa) / ((cn + fa) * (fa + hit) + (cn + miss) * (miss + hit)) * 100;
        case "ETS (Equitable Threat Score)":
            value = (hit - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) / ((hit + fa + miss) - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) * 100;
        default:
            return 0, fmt.Errorf("builder_stats.calculateStatCTC: ", "Invalid statistic:", statistic)
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
        case "RMSE":
            value = math.Sqrt(squareDiffSum / NSum);
            break;
        case "Bias (Model - Obs)":
            value = (modelSum - obsSum) / NSum;
            break;
        case "MAE (temp and dewpoint only)":
        case "MAE":
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
    var ctlLength int = len(data.ctlPop)
    var expLength int = len(data.expPop)
    var maxLen = int(math.Max(float64(ctlLength),float64(expLength)))
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

/*    subSecs = int64
    independentVarGroups = []
    independentVarHasPoint = []
    currIndependentVar
    curveIndex
    data
    di

    // matching in this function is based on a curve"s time variable which is an epoch.
    const independentVarName = "x"
    const statVarName = "y"

    // find the matching independentVars shared across all curves
    for (curveIndex = 0; curveIndex < 2; curveIndex++) {
        independentVarGroups[curveIndex] = [];  // array for the independentVars for each curve that are not null
        independentVarHasPoint[curveIndex] = [];   // array for the *all* of the independentVars for each curve
        subSecs[curveIndex] = {};  // map of the individual record times (subSecs) going into each independentVar for each curve
        data = dataset[curveIndex];
        // loop over every independentVar value in this curve
        for (di = 0; di < data[independentVarName].length; di++) {
            currIndependentVar = data[independentVarName][di];
            if (data[statVarName][di] !== null) {
                // store raw secs for this independentVar value, since it's not a null point
                subSecs[curveIndex][currIndependentVar] = data.subSecs[di];
                // store this independentVar value, since it's not a null point
                independentVarGroups[curveIndex].push(currIndependentVar);
            }
            // store all the independentVar values, regardless of whether they're null
            independentVarHasPoint[curveIndex].push(currIndependentVar);
        }
    }

    var matchingIndependentVars = _.intersection.apply(_, independentVarGroups);    // all of the non-null independentVar values common across all the curves
    var matchingIndependentHasPoint = _.intersection.apply(_, independentVarHasPoint);    // all of the independentVar values common across all the curves, regardless of whether or not they're null

    // remove non-matching independentVars and subSecs
    for (curveIndex = 0; curveIndex < 2; curveIndex++) { // loop over every curve
        data = dataset[curveIndex];
        // need to loop backwards through the data array so that we can splice non-matching indices
        // while still having the remaining indices in the correct order
        var dataLength = data[independentVarName].length;
        for (di = dataLength - 1; di >= 0; di--) {
            if (matchingIndependentVars.indexOf(data[independentVarName][di]) === -1) {
                // if this is not a common non-null independentVar value, we'll have to remove some data
                if (matchingIndependentHasPoint.indexOf(data[independentVarName][di]) === -1) {
                    // if at least one curve doesn't even have a null here, much less a matching value (because of the cadence), just drop this independentVar
                    matsDataUtils.removePoint(data, di, plotType, statVarName, isCTC, isScalar, hasLevels);
                } else {
                    // if all of the curves have either data or nulls at this independentVar, and there is at least one null, ensure all of the curves are null
                    matsDataUtils.nullPoint(data, di, statVarName, isCTC, isScalar, hasLevels);
                }
            }
        }

        dataset[curveIndex] = data;
    }
*/
    return dataset, nil
}

