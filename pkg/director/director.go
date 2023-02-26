package director

import (
	"fmt"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
)
var gp builder.GoodnessPolarity = 1
var minorThreshold builder.Threshold = 0.05
var majorThreshold builder.Threshold = 0.01
var ip builder.InputData
var cellBuilder ScorecardCellBuilder = builder.GetBuilder("TwoSampleTTest")

cellBuilder.SetGoodnessPolarity(gp)
cellBuilder.SetMinorThreshold(minorThreshold)
cellBuilder.SetMajorThreshold(majorThreshold)
cellBuilder.DeriveData(ip)
cell := cellBuilder.GetScorecardCell()
cellBuilder.ComputeSignificance(cell.Data)
value := cell.Value
