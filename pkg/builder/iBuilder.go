package builder

/*
please refer to https://gist.github.com/vaskoz/10073335 for an
example of a classic builder pattern.

Contents of a ScorecardCell
A ScorecardCell is a structure that represents a cell
in a scorecard display.

Each cell must have derivedData which is a
struct that has a control and an experimental array of floats.

Each cell must also have a GoodnessPolarity which is either a -1 or a 1
and defines whether a negative difference between the means
of the experimental array and the control array is better when
it is negative or better when it is positive.

Each cell must also have major and minor thresholds which help to define the value.

Each ScorecardCell must also have a resultant value.

A ScorecardCellBuilder is an interface that provides several functions:
	setGoodnessPolarity - sets the positive or negative direction of "goodness"
	setMajorThreshold - sets the major threshold
	setMinorThreshold - sets the minor threshold
	deriveData - creates DerivedData from InputData
	computeSignificance - calculates a value of a cell
	build() - converts an array of DerivedDataElements into
	an array of ScoredcardCell values (which are ints)

	An instance of a ScorecardCell struct implements a ScorecardCellBuilder
	interface by defining all the functions of the interface like...
	func (scc *ScorecardCell) setMajorThreshold(threshold Threshold) ScorecardCellBuilder {...}
	if ALL of the functions in the particular ScorecardCellBuilder are defined within a
	specific Builder then that builder can return a ScorecardCell like...
		func NewTwoSampleTTestBuilder() ScorecardCellBuilder {
			return &ScorecardCell{}
		}
    Then the instance of the new builder can execute all the functions
	of the builder to set the particular values, derive data, and cumpute
	significance values for an array of input data elements.
*/

type DerivedDataElement struct {
	CtlPop []float64
	ExpPop []float64
}

type DerivedData []DerivedDataElement

// -1 or 1
type GoodnessPolarity int
type Threshold float64
type ScorecardCell struct {
	Data             DerivedDataElement
	GoodnessPolarity GoodnessPolarity
	MajorThreshold   Threshold
	MinorThreshold   Threshold
	StatValue        float64
	Value            int
}

type ScorecardCellBuilder interface {
	SetGoodnessPolarity(GoodnessPolarity)
	SetMajorThreshold(Threshold)
	SetMinorThreshold(Threshold)
	SetInputData(DerivedDataElement)
	//DeriveData(InputData)
	ComputeSignificance(DerivedDataElement)
}

func GetBuilder(builderType string) *ScorecardCell {
	if builderType == "TwoSampleTTest" {
		return NewTwoSampleTTestBuilder()
	}
	return nil
}
