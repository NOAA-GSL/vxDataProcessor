package director

import (
	"fmt"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
)

var gp GoodnessPolarity = 1
var minorThreshold Threshold = 0.05
var majorThreshold Threshold = 0.01
var ip InputData
builderResult := NewTwoSampleTTestBuilder()
derivedData = DerivedData = builderResult.deriveData(ip)
builderResult.setGoodnessPolarity(gp).
	setMinorThreshold(minorThreshold).
	setMajorThreshold(majorThreshold).
	computeSignificance(derivedData)
