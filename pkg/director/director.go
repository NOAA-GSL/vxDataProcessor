package director

import (
	"fmt"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
)

var gp = builder.GoodnessPolarity(1)
var minorThreshold = builder.Threshold(0.05)
var majorThreshold = builder.Threshold(0.01)
var inputData = builder.DerivedDataElement{
	CtlPop: []float64{0.2, 1.3, 3.2, 4.5},
	ExpPop: []float64{0.1, 1.5, 3.0, 4.1},
}

func Build(documentId string) {
	// get the scorecard document
	// for all the input elements fire off a thread to do the compute
	var cellPtr = builder.GetBuilder("TwoSampleTTest")
	cellPtr.SetGoodnessPolarity(gp)
	cellPtr.SetMinorThreshold(minorThreshold)
	cellPtr.SetMajorThreshold(majorThreshold)
	cellPtr.SetInputData(inputData)
	cellPtr.ComputeSignificance(cellPtr.Data)
	var value = cellPtr.Value
	// insert the elements into the in-memory document
	fmt.Println(value)
	// upsert the document
}
