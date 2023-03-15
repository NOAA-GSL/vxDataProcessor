package builder_stats

import (
	"reflect"
	"testing"
	"time"
)

func getDataSet(epoch int64, ctlValues []float64, expValues []float64) DataSet {
	var ctlLen = len(ctlValues)
	var expLen = len(expValues)
	var tmpc = make([]PreCalcRecord, ctlLen)
	var tmpe = make([]PreCalcRecord, expLen)
	for i := 0; i < ctlLen; i++ {
		tmpc[i] = PreCalcRecord{time: epoch + int64(i), value: ctlValues[i]}
	}
	for i := 0; i < expLen; i++ {
		tmpe[i] = PreCalcRecord{time: epoch + int64(i), value: expValues[i]}
	}
	var dataSet = DataSet{
		ctlPop: tmpc,
		expPop: tmpe,
	}
	return dataSet
}
func TestGetMatchedDataSet(t *testing.T) {
	var epoch = time.Now().Unix()
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
			args: getDataSet(epoch, []float64{1.0, 2.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 4.0, 5.0}),
			want: getDataSet(epoch, []float64{1.0, 2.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 4.0, 5.0}),
			wantErr: false,
		},
		{
			name: "dataExpHole",
			args: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 5.0}),
			want: getDataSet(epoch, []float64{1.0, 2.0, 3.0, 5.0},
				[]float64{1.0, 2.0, 3.0, 5.0}),
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

func Test_calculateStatScalar(t *testing.T) {

	/*
	   Statistics for scalar
	   "RMSE" - surface
	   "Bias (Model - Obs)" - surface
	   "MAE (temp and dewpoint only)" - surface
	   "MAE" - surfrad

	   Associated test.sql files are in the test_data directory.
	   You can reproduce these test case numbers by using the associated app for
	   a statistic above and plugging in the values from the query and
	   plotting the time series curve. Get the earliest statistical value from the plot
	   (use the text output). Then get the statistical inputs for the test case by
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
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		//test cases.
		{
			//RMSE.sql
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
			want:    1.957,
			wantErr: false,
		},
		{
			//BIAS_MODEL_OBS.sql
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
			want:    -0.5741,
			wantErr: false,
		},
		{
			// MAE_temp_dewpoint.sql
			name: "MAE (temp and dewpoint only)",
			args: args{
				squareDiffSum:   22019.0390625,
				NSum:            1775,
				obsModelDiffSum: 1834.199951171875,
				modelSum:        85194.69848632812,
				obsSum:          87028.8984375,
				absSum:          87028.8984375,
				statistic:       "MAE (temp and dewpoint only)",
			},
			want:    1.530,
			wantErr: false,
		},
		{
			// MAE.sql
			name: "MAE",
			args: args{
				squareDiffSum:   3.5396907496750747,
				NSum:            13,
				obsModelDiffSum: -2.4978950321674347,
				modelSum:        0,
				obsSum:          -2.4978950321674347,
				absSum:          4.271478652954102,
				statistic:       "MAE",
			},
			want:    0.3286,
			wantErr: false,
		},
		// the following are error cases - don't need precise inputs
		{
			name: "MissingSquareDiffSum",
			args: args{
				//squareDiffSum:   1.0,
				NSum:            2.0,
				obsModelDiffSum: 3.0,
				modelSum:        4.0,
				obsSum:          5.0,
				absSum:          6.0,
				statistic:       "squareDiffSum",
			},
			want:    0.0,
			wantErr: true,
		},
		{
			name: "MissingObsModelDiffSum",
			args: args{
				squareDiffSum: 0.0,
				NSum:          2.0,
				//obsModelDiffSum: 3.0,
				modelSum:  4.0,
				obsSum:    5.0,
				absSum:    6.0,
				statistic: "obsModelDiffSum",
			},
			want:    0.0,
			wantErr: true,
		},
		{
			name: "MissingNSum",
			args: args{
				squareDiffSum: 1.0,
				//NSum:            2.0,
				obsModelDiffSum: 3.0,
				modelSum:        4.0,
				obsSum:          5.0,
				absSum:          6.0,
				statistic:       "NSum",
			},
			want:    0.0,
			wantErr: true,
		},
		{
			name: "MissingModelSum",
			args: args{
				squareDiffSum:   1.0,
				NSum:            2.0,
				obsModelDiffSum: 3.0,
				//modelSum:        4.0,
				obsSum:    5.0,
				absSum:    6.0,
				statistic: "modelSum",
			},
			want:    0.0,
			wantErr: true,
		},
		{
			name: "MissingObsSum",
			args: args{
				squareDiffSum:   1.0,
				NSum:            2.0,
				obsModelDiffSum: 3.0,
				modelSum:        4.0,
				//obsSum:          5.0,
				absSum:    6.0,
				statistic: "obsSum",
			},
			want:    0.0,
			wantErr: true,
		},
		{
			name: "MissingAbsSum",
			args: args{
				squareDiffSum:   1.0,
				NSum:            2.0,
				obsModelDiffSum: 3.0,
				modelSum:        4.0,
				obsSum:          5.0,
				//absSum:          6.0,
				statistic: "absSum",
			},
			want:    0.0,
			wantErr: true,
		},
		{
			name: "MissingStatistic",
			args: args{
				squareDiffSum:   1.0,
				NSum:            2.0,
				obsModelDiffSum: 3.0,
				modelSum:        4.0,
				obsSum:          5.0,
				absSum:          6.0,
				//statistic:       "",
			},
			want:    0.0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateStatScalar(tt.args.squareDiffSum, tt.args.NSum, tt.args.obsModelDiffSum, tt.args.modelSum, tt.args.obsSum, tt.args.absSum, tt.args.statistic)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateStatScalar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateStatScalar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateStatCTC(t *testing.T) {
	/*
	   statistics for CTC
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
	   a statistic above and plugging in the values from the query and
	   plotting the time series curve. Get the earliest statistical value from the plot
	   (use the text output). Then get the statistical inputs for the test case by
	   running the associated test_data/{stat}.sql query.
	*/

	type args struct {
		hit       int
		fa        int
		miss      int
		cn        int
		statistic string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		//test cases.
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
			//PODy_lt.sql - ceiling
			name: "PODy (POD of value < threshold)",
			args: args{
				hit:       10,
				fa:        46,
				miss:      18,
				cn:        1695,
				statistic: "PODy (POD of value < threshold)",
			},
			want:    	35.71,
			wantErr: false,
		},
		{
			//PODy_gt.sql - radar
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
			//PODn_gt.sql - ceiling
			name: "PODn (POD of value > threshold)",
			args: args{
				hit:       10,
				fa:        46,
				miss:      18,
				cn:        1695,
				statistic: "PODn (POD of value > threshold)",
			},
			want:    	97.36,
			wantErr: false,
		},
		{
			//PODn_lt.sql - radar
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
			//FAR.sql - radar
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
			//CSI.sql - radar
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
			//HSS.sql - radar
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
			//ETS.sql - radar
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
			name: "MissingHit",
			args: args{
				//hit:   		1,
				fa:        2,
				miss:      3,
				cn:        4,
				statistic: "TSS (True Skill Score)",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "MissingFa",
			args: args{
				hit: 1,
				//fa:    		2,
				miss:      3,
				cn:        4,
				statistic: "TSS (True Skill Score)",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "MissingMiss",
			args: args{
				hit: 1,
				fa:  2,
				//miss:  		3,
				cn:        4,
				statistic: "TSS (True Skill Score)",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "MissingCn",
			args: args{
				hit:  1,
				fa:   2,
				miss: 3,
				//cn:    		4,
				statistic: "TSS (True Skill Score)",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateStatCTC(tt.args.hit, tt.args.fa, tt.args.miss, tt.args.cn, tt.args.statistic)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateStatCTC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateStatCTC() = %v, want %v", got, tt.want)
			}
		})
	}
}
