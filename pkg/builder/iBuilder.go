package builder

/*
Contents of a ScorecardCell
A ScorecardCell is a structure that represents a cell
in a scorecard display.

Each cell must have derivedData which is a
struct that has two arrays of numbers, a control and an experimental.

Each cell must also have a GoodnessPolarity which is either a -1 or a 1
and defines whether a negative difference between the means
of the experimental array and the control array is better when
it is negative or better when it is positive.

Each cell must also have major and minor thresholds which define
the confidence thresholds against which the statistical  probability value will be compared.

Each ScorecardCell must also have a resultant value pointer. This pointer points to
the result location into which the computeSignificance will write the result.

A ScorecardCellBuilder is an interface that provides several functions:
	setGoodnessPolarity - sets the positive or negative direction of "goodness"
	setMajorThreshold - sets the major threshold
	setMinorThreshold - sets the minor threshold
	deriveInputData - creates DerivedData from InputData. This requires that the function
	performs time matching on the input populations, then performs a statistic calculation,
	and then writes the DerivedDataElement into the scorecardCell Data.
	computeSignificance - calculates and writes the value of a cell.

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

type statFunc func(StatType)
type StatType string

type DerivedDataElement struct {
	CtlPop []float64
	ExpPop []float64
}

type DerivedData []DerivedDataElement

// -1 or 1
type GoodnessPolarity int
type Threshold float64
type ScorecardCell struct {
	data             DerivedDataElement
	goodnessPolarity GoodnessPolarity
	majorThreshold   Threshold
	minorThreshold   Threshold
	pvalue           float64
	valuePtr            *int
}


type QueryResult = *[]struct{} // either CTCRecord or ScalarRecord or PreCalcRecord
type ScorecardCellBuilder interface {
	SetGoodnessPolarity(GoodnessPolarity)
	SetMajorThreshold(Threshold)
	SetMinorThreshold(Threshold)
	DeriveInputData(input QueryResult, statisticType string)
	ComputeSignificance(scc *ScorecardCell)
	SetValuePtr(valuePtr *int)
	GetValue()
	Build(cellPtr *ScorecardCell, inputData struct{})
}

func GetBuilder(builderType string) *ScorecardCell {
	if builderType == "TwoSampleTTest" {
		return NewTwoSampleTTestBuilder()
	}
	return nil
}
