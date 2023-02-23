package builder_test

import (
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"testing"
)

func TestTwoSampleTTestBuilder(t *testing.T) {
	var gp GoodnessPolarity = 1
	var minorThreshold Threshold = 0.05
	var majorThreshold Threshold = 0.01
	// type InputData struct {
	// 	Rows []struct {
	// 		InputData []DerivedData
	// 	}
	// }
	var ip = InputData{
		[3]{DerivedData{}, DerivedData{}, DerivedData{}}
	}
	var builderResult BuilderResult = NewTwoSampleTTestBuilder()
	var derivedData DerivedData = builderResult.deriveData(ip)
	builderResult.setGoodnessPolarity(gp).
		setMinorThreshold(minorThreshold).
		setMajorThreshold(majorThreshold).
		computeSignificance(derivedData)
}
