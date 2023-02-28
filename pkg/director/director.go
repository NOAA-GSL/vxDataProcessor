package director

import (
	"fmt"
	"log"
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
	err := cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetGoodnessPolarity - error message : ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(inputData)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - ComputeSignificance - error message : ", err))
	}
	// insert the elements into the in-memory document
	// upsert the document
}
