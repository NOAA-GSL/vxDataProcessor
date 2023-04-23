package builder

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getDataSet(epoch int64, ctlValues []float64, expValues []float64) DataSet {
	ctlLen := len(ctlValues)
	expLen := len(expValues)
	tmpc := make([]PreCalcRecord, ctlLen)
	tmpe := make([]PreCalcRecord, expLen)
	for i := 0; i < ctlLen; i++ {
		tmpc[i] = PreCalcRecord{Avtime: epoch + int64(ctlValues[i]), Stat: ctlValues[i]}
	}
	for i := 0; i < expLen; i++ {
		tmpe[i] = PreCalcRecord{Avtime: epoch + int64(expValues[i]), Stat: expValues[i]}
	}
	dataSet := DataSet{
		ctlPop: tmpc,
		expPop: tmpe,
	}
	return dataSet
}

func TestGetMatchedDataSet(t *testing.T) {
	epoch := time.Now().Unix()
	tests := []struct {
		name    string
		args    DataSet
		want    DataSet
		wantErr bool
	}{
		// test cases.
		{
			name: "matchedData",
			args: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0}),
			want: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0}),
			wantErr: false,
		},
		{
			name: "dataCtlHole",
			args: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 4.0, 5.0}),
			want: getDataSet(epoch, []float64{1.0, 2.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 4.0, 5.0}),
			wantErr: false,
		},
		{
			name: "dataExpHole",
			args: getDataSet(epoch, []float64{1.0, 2.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0}),
			want: getDataSet(epoch, []float64{1.0, 2.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 4.0, 5.0}),
			wantErr: false,
		},
		{
			name: "dataFirstHole",
			args: getDataSet(epoch, []float64{2.0, 3.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0}),
			want: getDataSet(epoch, []float64{2.0, 3.0, 4.0, 5.0},
				[]float64{2.0, 3.0, 4.0, 5.0}),
			wantErr: false,
		},
		{
			name: "dataLastHole",
			args: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 4.0},
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0}),
			want: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 4.0},
				[]float64{1.0, 2.0, 3.0, 4.0}),
			wantErr: false,
		},
		{
			name: "dataTwoHoles",
			args: getDataSet(epoch, []float64{1.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0}),
			want: getDataSet(epoch, []float64{1.0, 4.0, 5.0},
				[]float64{1.0, 4.0, 5.0}),
			wantErr: false,
		},
		{
			name: "dataAllHoles",
			args: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				[]float64{}),
			want: getDataSet(epoch, []float64{},
				[]float64{}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMatchedDataSet(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMatchedDataSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMatchedDataSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// this test has inputs captured from a real world example
func TestGetMatchedDataSetRealWorld(t *testing.T) {
	var ctlData, expData PreCalcRecords
	var dataSet DataSet
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678788000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678791600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678795200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678798800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678802400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678806000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678809600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678813200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678816800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678820400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678824000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678827600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678831200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678834800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678838400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678842000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678845600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678849200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678852800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678856400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678860000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678863600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678867200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678870800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678874400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678878000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678881600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678885200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678888800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678892400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678896000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678899600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678903200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678906800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678910400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678914000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678917600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678921200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678924800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678928400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678932000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678935600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678939200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678942800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678946400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678950000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678953600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678957200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678960800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678964400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678968000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678971600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678975200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678978800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678982400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678986000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678989600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678993200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1678996800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679000400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679004000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679007600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679011200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679014800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679018400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679022000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679025600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679029200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679032800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679036400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679040000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679043600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679047200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679050800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679054400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679058000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679061600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679065200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679068800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679072400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679076000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679079600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679083200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679086800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679090400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679094000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679097600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679101200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679104800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679108400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679112000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679115600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679119200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679122800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679126400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679130000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679133600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679137200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679140800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679144400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679148000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679151600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679155200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679158800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679162400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679166000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679169600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679173200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679176800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679180400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679184000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679187600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679191200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679194800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679198400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679202000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679205600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679209200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679212800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679216400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679220000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679223600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679227200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679230800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679234400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679238000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679241600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679245200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679248800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679252400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679256000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679259600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679263200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679266800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679270400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679274000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679277600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679281200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679284800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679288400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679292000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679295600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679299200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679302800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679306400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679310000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679313600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679317200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679320800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679324400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679328000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679331600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679335200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679338800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679342400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679346000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679349600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679353200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679356800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679360400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679364000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679367600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679371200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679374800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679378400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679382000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679385600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679389200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679392800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679396400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679400000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679403600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679407200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679410800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679414400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679418000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679421600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679425200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679428800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679432400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679436000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679439600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679443200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679446800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679450400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679454000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679457600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679461200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679464800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679468400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679472000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679475600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679479200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679482800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679486400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679490000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679493600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679497200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679500800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679504400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679508000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679511600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679515200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679518800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679533200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679536800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679540400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679544000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679547600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679551200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679554800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679558400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679562000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679565600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679569200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679572800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679576400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679580000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679583600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679587200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679590800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679594400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679598000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679601600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679605200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679608800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679612400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679616000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679619600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679623200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679626800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679630400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679634000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679637600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679641200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679644800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679648400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679652000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679655600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679659200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679662800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679666400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679670000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679673600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679677200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679680800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679684400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679688000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679691600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679695200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679698800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679702400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679706000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679709600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679713200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679716800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679720400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679724000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679727600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679731200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679734800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679738400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679742000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679745600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679749200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679752800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679756400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679760000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679763600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679767200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679770800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679774400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679778000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679781600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679785200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679788800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679792400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679796000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679799600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679803200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679806800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679810400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679814000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679817600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679821200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679824800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679828400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679832000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679835600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679839200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679842800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679846400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679850000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679853600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679857200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679860800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679864400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679868000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679871600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679875200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679878800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679882400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679886000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679889600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679893200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679896800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679900400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679904000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679907600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679911200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679914800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679918400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679922000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679925600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679929200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679932800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679936400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679940000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679943600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679947200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679950800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679954400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679958000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679961600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679965200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679968800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679972400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679976000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679979600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679983200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679986800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679990400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679994000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1679997600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680001200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680004800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680008400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680012000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680015600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680019200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680022800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680026400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680030000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680033600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680037200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680040800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680044400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680048000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680051600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680055200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680058800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680062400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680066000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680069600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680073200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680076800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680080400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680084000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680087600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680091200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680094800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680098400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680102000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680105600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680109200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680112800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680116400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680120000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680123600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680127200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680130800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680134400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680138000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680141600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680145200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680148800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680152400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680156000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680159600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680163200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680166800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680170400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680174000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680177600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680181200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680184800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680188400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680192000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680195600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680199200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680202800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680206400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680210000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680213600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680217200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680220800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680224400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680228000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680231600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680235200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680238800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680242400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680246000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680249600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680253200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680256800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680260400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680264000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680267600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680271200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680274800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680278400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680282000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680285600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680289200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680292800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680296400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680300000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680303600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680307200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680310800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680314400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680318000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680321600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680325200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680328800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680332400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680336000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680339600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680343200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680346800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680350400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680354000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680357600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680361200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680364800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680368400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680372000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680375600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680379200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680382800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680386400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680390000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680393600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680397200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680400800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680404400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680408000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680411600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680415200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680418800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680422400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680426000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680429600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680433200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680436800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680440400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680444000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680447600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680451200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680454800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680458400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680462000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680465600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680469200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680472800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680476400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680480000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680483600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680487200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680490800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680494400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680498000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680501600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680505200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680508800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680512400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680516000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680519600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680523200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680526800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680530400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680534000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680537600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680541200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680544800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680548400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680552000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680555600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680559200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680562800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680566400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680570000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680573600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680577200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680580800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680584400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680588000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680591600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680595200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680598800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680602400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680606000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680609600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680613200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680616800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680620400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680624000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680627600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680631200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680634800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680638400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680642000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680645600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680649200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680652800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680656400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680660000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680663600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680667200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680670800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680674400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680678000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680681600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680685200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680688800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680692400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680696000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680699600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680703200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680706800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680710400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680714000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680717600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680721200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680724800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680728400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680732000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680735600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680739200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680742800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680746400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680750000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680753600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680757200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680760800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680764400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680768000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680771600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680775200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680778800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680782400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680786000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680789600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680793200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680796800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680800400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680804000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680807600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680811200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680814800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680818400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680822000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680825600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680829200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680832800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680836400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680840000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680843600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680847200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680850800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680854400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680858000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680861600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680865200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680868800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680872400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680876000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680879600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680883200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680886800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680890400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680894000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680897600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680901200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680904800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680908400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680912000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680915600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680919200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680922800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680926400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680930000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680933600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680937200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680940800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680944400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680948000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680951600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680955200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680958800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680962400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680966000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680969600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680973200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680976800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680980400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680984000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680987600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680991200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680994800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1680998400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681002000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681005600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681009200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681012800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681016400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681020000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681023600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681027200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681030800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681034400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681038000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681041600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681045200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681048800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681052400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681056000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681059600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681063200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681066800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681070400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681074000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681077600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681081200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681084800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681088400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681092000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681095600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681099200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681102800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681106400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681110000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681113600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681117200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681120800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681124400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681128000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681131600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681135200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681138800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681142400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681146000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681149600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681153200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681156800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681160400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681164000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681167600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681171200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681174800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681178400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681182000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681185600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681203600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681207200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681210800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681214400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681218000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681221600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681225200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681228800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681232400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681236000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681239600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681243200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681246800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681250400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681254000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681257600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681261200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681264800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681268400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681272000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681275600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681279200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681282800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681286400, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681290000, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681293600, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681297200, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681300800, Stat: 0})
	ctlData = append(ctlData, PreCalcRecord{Avtime: 1681304400, Stat: 0})

	expData = append(expData, PreCalcRecord{Avtime: 1678773600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1678816800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1678903200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1678946400, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679032800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679076000, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679119200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679162400, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679205600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679248800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679292000, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679335200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679378400, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679421600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679464800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679551200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679594400, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679637600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679680800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679724000, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679767200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679810400, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1679853600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680069600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680112800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680156000, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680199200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680242400, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680285600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680588000, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680717600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680760800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680804000, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680847200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680933600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1680976800, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1681020000, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1681063200, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1681106400, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1681149600, Stat: 0})
	expData = append(expData, PreCalcRecord{Avtime: 1681192800, Stat: 0})

	dataSet.ctlPop = ctlData
	dataSet.expPop = expData

	got, err := GetMatchedDataSet(dataSet)
	if err != nil {
		t.Errorf("GetMatchedDataSet() error = %v", err)
	}
	if len(got.ctlPop) != 39 {
		t.Errorf("GetMatchedDataSetRealWorld() len(got.ctlPop)= %v, wanted %v", len(got.ctlPop), 41)
	}
	if len(got.expPop) != 39 {
		t.Errorf("GetMatchedDataSetRealWorld() len(got.expPop)= %v, wanted %v", len(got.expPop), 41)
	}
}

func Test_calculateStatScalar(t *testing.T) {
	/*
	  , Statistics for scalar
	   "RMSE" - surface
	   "Bias (Model - Obs)" - surface
	   "MAE (temp and dewpoint only)" - surface
	   "MAE" - surfrad

	   Associated test.sql files are in the test_data directory.
	   You can reproduce these test case numbers by using the associated app for
	   a, Statistic above and plugging in the values from the query and
	   plotting the time series curve. Get the earliest, Statistical value from the plot
	   (use the text output). Then get the, Statistical inputs for the test case by
	   running the associated test_data/{stat}.sql query.
	*/

	type args struct {
		squareDiffSum   float64
		NSum            float64
		obsModelDiffSum float64
		modelSum        float64
		obsSum          float64
		absSum          float64
		statistic       string
	}
	tests := []struct {
		name      string
		args      args
		want      float64
		tolerance float64
		wantErr   bool
	}{
		// test cases.
		{
			// RMSE.sql
			name: "RMSE",
			args: args{
				squareDiffSum:   22019.0390625,
				NSum:            1775,
				obsModelDiffSum: 1834.199951171875,
				modelSum:        85194.69848632812,
				obsSum:          87028.8984375,
				absSum:          4889.7998046875,
				statistic:       "RMSE",
			},
			want:      1.957 * 1.8,
			tolerance: 0.005,
			wantErr:   false,
		},
		{
			// BIAS_MODEL_OBS.sql
			name: "Bias (Model - Obs)",
			args: args{
				squareDiffSum:   22019.0390625,
				NSum:            1775,
				obsModelDiffSum: 1834.199951171875,
				modelSum:        85194.69848632812,
				obsSum:          87028.8984375,
				absSum:          4889.7998046875,
				statistic:       "Bias (Model - Obs)",
			},
			want:      -0.5741 * 1.8,
			tolerance: 0.001,
			wantErr:   false,
		},
		{
			// MAE_temp_dewpoint.sql
			name: "MAE (temp and dewpoint only)",
			args: args{
				squareDiffSum:   4328.60986328125,
				NSum:            212,
				obsModelDiffSum: 67.5,
				modelSum:        6096.10009765625,
				obsSum:          6163.60009765625,
				absSum:          740.9000244140630,
				statistic:       "MAE (temp and dewpoint only)",
			},
			want:      1.942 * 1.8,
			tolerance: 0.005,
			wantErr:   false,
		},
		{
			// MAE.sql
			name: "MAE",
			args: args{
				squareDiffSum:   3.5396907496750747,
				NSum:            13,
				obsModelDiffSum: -2.4978950321674347,
				modelSum:        0.1,
				obsSum:          -2.4978950321674347,
				absSum:          4.271478652954102,
				statistic:       "MAE",
			},
			want:      0.3286,
			tolerance: 0.005,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got float64
			var err error
			got, err = CalculateStatScalar(tt.args.squareDiffSum, tt.args.NSum, tt.args.obsModelDiffSum, tt.args.modelSum, tt.args.obsSum, tt.args.absSum, tt.args.statistic)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateStatScalar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("calculateStatScalar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateStatCTC(t *testing.T) {
	/*
	  , Statistics for CTC
	   "TSS (True Skill Score)" - radar
	   "PODy (POD of value < threshold)" - ceiling
	   "PODy (POD of value > threshold)" - radar
	   "PODn (POD of value > threshold)" - ceiling
	   "PODn (POD of value < threshold)" - radar
	   "FAR (False Alarm Ratio)" - radar
	   "CSI (Critical Success Index)" - radar
	   "HSS (Heidke Skill Score)" - radar
	   "ETS (Equitable Threat Score)" - radar

	   Associated test.sql files are in the test_data directory.
	   You can reproduce these test case numbers by using the associated app for
	   a, Statistic above and plugging in the values from the query and
	   plotting the time series curve. Get the earliest, Statistical value from the plot
	   (use the text output). Then get the, Statistical inputs for the test case by
	   running the associated test_data/{stat}.sql query.
	*/

	type args struct {
		hit       float32
		fa        float32
		miss      float32
		cn        float32
		statistic string
	}
	tests := []struct {
		name    string
		args    args
		want    float32
		wantErr bool
	}{
		// test cases.
		{
			// TSS.sql - radar
			name: "TSS (True Skill Score)",
			args: args{
				hit:       1583,
				fa:        1876,
				miss:      868,
				cn:        56054,
				statistic: "TSS (True Skill Score)",
			},
			want:    61.35,
			wantErr: false,
		},
		{
			// PODy_lt.sql - ceiling
			name: "PODy (POD of value < threshold)",
			args: args{
				hit:       10,
				fa:        46,
				miss:      18,
				cn:        1695,
				statistic: "PODy (POD of value < threshold)",
			},
			want:    35.71,
			wantErr: false,
		},
		{
			// PODy_gt.sql - radar
			name: "PODy (POD of value > threshold)",
			args: args{
				hit:       1583,
				fa:        1876,
				miss:      868,
				cn:        56054,
				statistic: "PODy (POD of value > threshold)",
			},
			want:    64.59,
			wantErr: false,
		},
		{
			name: "PODn (POD of value > threshold)",
			args: args{
				hit:       2,
				fa:        41,
				miss:      5,
				cn:        1716,
				statistic: "PODn (POD of value > threshold)",
			},
			want:    97.67,
			wantErr: false,
		},
		{
			// PODn_lt.sql - radar
			name: "PODn (POD of value < threshold)",
			args: args{
				hit:       1583,
				fa:        1876,
				miss:      868,
				cn:        56054,
				statistic: "PODn (POD of value < threshold)",
			},
			want:    96.76,
			wantErr: false,
		},
		{
			// FAR.sql - radar
			name: "FAR (False Alarm Ratio)",
			args: args{
				hit:       1583,
				fa:        1876,
				miss:      868,
				cn:        56054,
				statistic: "FAR (False Alarm Ratio)",
			},
			want:    54.24,
			wantErr: false,
		},
		{
			// CSI.sql - radar
			name: "CSI (Critical Success Index)",
			args: args{
				hit:       1583,
				fa:        1876,
				miss:      868,
				cn:        56054,
				statistic: "CSI (Critical Success Index)",
			},
			want:    36.58,
			wantErr: false,
		},
		{
			// HSS.sql - radar
			name: "HSS (Heidke Skill Score)",
			args: args{
				hit:       1583,
				fa:        1876,
				miss:      868,
				cn:        56054,
				statistic: "HSS (Heidke Skill Score)",
			},
			want:    51.25,
			wantErr: false,
		},
		{
			// ETS.sql - radar
			name: "ETS (Equitable Threat Score)",
			args: args{
				hit:       1583,
				fa:        1876,
				miss:      868,
				cn:        56054,
				statistic: "ETS (Equitable Threat Score)",
			},
			want:    34.46,
			wantErr: false,
		},
		{
			// TSS.sql - radar - Not A Number error
			name: "Not a Number",
			args: args{
				hit:       0,
				fa:        1876,
				miss:      0,
				cn:        56054,
				statistic: "TSS (True Skill Score)",
			},
			want:    -9999,
			wantErr: false,
		},
		{
			// TSS.sql - radar - infinity
			// don't know how to cause this condition with valid params
			name: "infinity",
			args: args{
				hit:       1,
				fa:        1,
				miss:      1,
				cn:        1,
				statistic: "TSS (True Skill Score)",
			},
			want: 0.0,
			// wantErr: true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var delta float64 = 0.006
			got, err := CalculateStatCTC(tt.args.hit, tt.args.fa, tt.args.miss, tt.args.cn, tt.args.statistic)
			if tt.wantErr {
				assert.Errorf(t, err, "calculateStatCTC() should have returned error but did not - got %w", got)
			} else {
				assert.NoErrorf(t, err, "calculateStatCTC() returned error %w", err)
				assert.InDelta(t, tt.want, got, delta, "calculateStatCTC() excessive difference")
			}
		})
	}
}
