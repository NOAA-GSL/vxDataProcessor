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
import (
	"sync"
)

type StatType string

const ErrorValue = -9999

type DerivedDataElement struct {
	CtlPop []float64
	ExpPop []float64
}

type (
	GoodnessPolarity int // -1 or 1
	Threshold        float64
	ScorecardCell    struct {
		mu               sync.Mutex
		Data             DerivedDataElement
		goodnessPolarity GoodnessPolarity
		majorThreshold   Threshold
		minorThreshold   Threshold
		statisticType    StatisticType
		pvalue           float64
		keychain         []string
		value            int
	}
)

// these are floats because of the division in the CalculateStatCTC func
type CTCRecord struct {
	Avtime int64
	Hit    float32
	Miss   float32
	Fa     float32
	Cn     float32
}
type CTCRecords = []CTCRecord

type ScalarRecord struct {
	Avtime          int64
	SquareDiffSum   float64
	NSum            float64
	ObsModelDiffSum float64
	ModelSum        float64
	ObsSum          float64
	AbsSum          float64
}
type ScalarRecords []ScalarRecord

type PreCalcRecord struct {
	Avtime int64
	Stat   float64
}
type PreCalcRecords []PreCalcRecord

type BuilderScalarResult struct {
	CtlData ScalarRecords
	ExpData ScalarRecords
}
type BuilderCTCResult struct {
	CtlData CTCRecords
	ExpData CTCRecords
}
type BuilderPreCalcResult struct {
	CtlData PreCalcRecords
	ExpData PreCalcRecords
}

// enum values for statistics type
type StatisticType int

const (
	TSS_True_Skill_Score StatisticType = iota
	PODy_POD_of_value_lt_threshold
	PODy_POD_of_value_gt_threshold
	PODn_POD_of_value_gt_threshold
	PODn_POD_of_value_lt_threshold
	FAR_False_Alarm_Ratio
	CSI_Critical_Success_Index
	HSS_Heidke_Skill_Score
	ETS_Equitable_Threat_Score
	ACC
	RMSE
	Bias_Model_Obs
	MAE_temp_and_dewpoint_only
	MAE
	Unknown
)

// implement the String interface for StatisticType
func (s StatisticType) String() string {
	switch s {
	case TSS_True_Skill_Score:
		return "TSS (True Skill Score)"
	case PODy_POD_of_value_lt_threshold:
		return "PODy (POD of value < threshold)"
	case PODy_POD_of_value_gt_threshold:
		return "PODy (POD of value > threshold)"
	case PODn_POD_of_value_gt_threshold:
		return "PODn (POD of value > threshold)"
	case PODn_POD_of_value_lt_threshold:
		return "PODn (POD of value < threshold)"
	case FAR_False_Alarm_Ratio:
		return "FAR (False Alarm Ratio)"
	case CSI_Critical_Success_Index:
		return "CSI (Critical Success Index)"
	case HSS_Heidke_Skill_Score:
		return "HSS (Heidke Skill Score)"
	case ETS_Equitable_Threat_Score:
		return "ETS (Equitable Threat Score)"
	case ACC:
		return "Want experimental to exceed control"
	case RMSE:
		return "RMSE"
	case Bias_Model_Obs:
		return "Bias (Model - Obs)"
	case MAE_temp_and_dewpoint_only:
		return "MAE (temp and dewpoint only)"
	case MAE:
		return "MAE"
	default:
		return "Unknown"
	}
}

// implment the reverse string interface for StatisticType
func GetStatisticTpe(statType string) StatisticType {
	switch statType {
	case "TSS (True Skill Score)":
		return TSS_True_Skill_Score
	case "PODy (POD of value < threshold)":
		return PODy_POD_of_value_lt_threshold
	case "PODy (POD of value > threshold)":
		return PODy_POD_of_value_gt_threshold
	case "PODn (POD of value > threshold)":
		return PODn_POD_of_value_gt_threshold
	case "PODn (POD of value < threshold)":
		return PODn_POD_of_value_lt_threshold
	case "FAR (False Alarm Ratio)":
		return FAR_False_Alarm_Ratio
	case "CSI (Critical Success Index)":
		return CSI_Critical_Success_Index
	case "HSS (Heidke Skill Score)":
		return HSS_Heidke_Skill_Score
	case "ETS (Equitable Threat Score)":
		return ETS_Equitable_Threat_Score
	case "ACC":
		return ACC
	case "RMSE":
		return RMSE
	case "Bias (Model - Obs)":
		return Bias_Model_Obs
	case "MAE (temp and dewpoint only)":
		return MAE_temp_and_dewpoint_only
	case "MAE":
		return MAE
	default:
		return Unknown
	}
}

type ScorecardCellBuilder interface {
	setGoodnessPolarity(GoodnessPolarity)
	setMajorThreshold(Threshold)
	setMinorThreshold(Threshold)
	SetKeyChain([]string) // has to be public
	deriveInputData(QueryResult interface{})
	computeSignificance()
	getValue()
	setValue(value int32)
	SetStatisticType(statisticType string)
	Build(qrPtr interface{}, statisticType string, minorThreshold float64, majorThreshold float64)
}

func GetBuilder(builderType string) *ScorecardCell {
	if builderType == "TwoSampleTTest" {
		return NewTwoSampleTTestBuilder()
	}
	return nil
}
