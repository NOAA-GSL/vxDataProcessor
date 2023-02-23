package builder

/*
please refer to https://gist.github.com/vaskoz/10073335
*/

type DerivedData struct {
	CtlPop []float64
	ExpPop []float64
}
type InputData struct {
	Rows []struct {
		InputData []DerivedData
	}
}

// -1 or 1
type GoodnessPolarity int
type Threshold float64
type builderResult struct {
	data             DerivedData
	goodnessPolarity GoodnessPolarity
	majorThreshold   Threshold
	minorThreshold   Threshold
	value            int
}

type Builder interface {
	deriveData(InputData) Builder
	setGoodnessPolarity(GoodnessPolarity) Builder
	setMajorThreshold(Threshold) Builder
	setMinorThreshold(Threshold) Builder
	computeSignificance(DerivedData) Builder
}
