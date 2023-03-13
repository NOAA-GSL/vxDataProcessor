package builder_stats


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
	NSum int
	obsModelDiffSum float64
	modelSum float64
	obsSum float64
	absSum float64
	time int64
}
type PreCalcRecord struct {
	value float64
	time int64
}

type DataSet struct{
    ctlPop []float64
    expPop []float64
}

/*
These are stats functions that are used to derive scorecard stats from raw populations.
There is also a time matching function. These functions are used by the builder functions.
*/

// calculates the statistic for ctc plots
func calculateStatCTC(hit int, fa int, miss int, cn int, statistic string) ([]float64, error){
/*    if isNaN(hit) || isNaN(fa) || isNaN(miss) || isNaN(cn) return nil;
        var queryVal;
        switch (statistic) {
        case 'TSS (True Skill Score)':
            queryVal = ((hit * cn - fa * miss) / ((hit + miss) * (fa + cn))) * 100;
            break;
        // some PODy measures look for a value over a threshold, some look for under
        case 'PODy (POD of value < threshold)':
        case 'PODy (POD of value > threshold)':
            queryVal = hit / (hit + miss) * 100;
            break;
        // some PODn measures look for a value under a threshold, some look for over
        case 'PODn (POD of value > threshold)':
        case 'PODn (POD of value < threshold)':
            queryVal = cn / (cn + fa) * 100;
            break;
        case 'POFD (Probability of False Detection)':
            queryVal = fa / (fa + cn) * 100;
            break;
        case 'FAR (False Alarm Ratio)':
            queryVal = fa / (fa + hit) * 100;
            break;
        case 'Bias (forecast/actual)':
            queryVal = (hit + fa) / (hit + miss);
            break;
        case 'CSI (Critical Success Index)':
            queryVal = hit / (hit + miss + fa) * 100;
            break;
        case 'HSS (Heidke Skill Score)':
            queryVal = 2 * (cn * hit - miss * fa) / ((cn + fa) * (fa + hit) + (cn + miss) * (miss + hit)) * 100;
            break;
        case 'ETS (Equitable Threat Score)':
            queryVal = (hit - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) / ((hit + fa + miss) - ((hit + fa) * (hit + miss) / (hit + fa + miss + cn))) * 100;
            break;
    }
    return queryVal;
*/
return nil, nil
}

// calculates the statistic for scalar partial sums plots
func calculateStatScalar (squareDiffSum float64, NSumfloat64, obsModelDiffSumfloat64, modelSumfloat64, obsSumfloat64, absSumfloat64, statistic string)([]float64, error) {
/*    if (isNaN(squareDiffSum) || isNaN(NSum) || isNaN(obsModelDiffSum) || isNaN(modelSum) || isNaN(obsSum) || isNaN(absSum)) return null;
    var queryVal;
    switch (statistic) {
        case 'RMSE':
            queryVal = Math.sqrt(squareDiffSum / NSum);
            break;
        case 'Bias (Model - Obs)':
            queryVal = (modelSum - obsSum) / NSum;
            break;
        case 'MAE (temp and dewpoint only)':
        case 'MAE':
            queryVal = absSum / NSum;
            break;
    }
    if (isNaN(queryVal)) return null;
    return queryVal;
    */
    return nil, nil
}

// function for removing unmatched data from a dataset containing multiple curves
func GetMatchedDataSet(InputData DataSet)([]float64, error){
    /*
    subSecs = []
    independentVarGroups = []
    independentVarHasPoint = []
    currIndependentVar
    curveIndex
    data
    di

    // matching in this function is based on a curve's independent variable. For a timeseries, the independentVar is epoch,
    //determine whether data.x or data.y is the independent variable, and which is the stat value
    const independentVarName = 'x'
    const statVarName = 'y'

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

    return dataset;
    */
    return nil, nil
}

