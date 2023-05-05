//go:build integration
// +build integration

package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/director"
	"github.com/couchbase/gocb/v2"
	"github.com/joho/godotenv"
	"go.uber.org/goleak"
)

func loadEnvironmentFile() {
	environmentFile, set := os.LookupEnv("PROC_ENV_PATH")
	if !set {
		err := godotenv.Load() // Loads from "$(pwd)/.env"
		if err != nil {
			log.Printf("Couldn't load environment file: %q", environmentFile)
		}
	} else {
		err := godotenv.Load(environmentFile) // Loads from whatever PROC_ENV_PATH has been set to
		if err != nil {
			log.Printf("Couldn't load environment file: %q", environmentFile)
		}
	}
}

func getTestDoc(mngr *Manager) (map[string]interface{}, error) {
	loadEnvironmentFile()
	// get the test scorecard document (this is a Result - not a document)
	var scorecardDataIn *gocb.GetResult
	scorecardDataIn, err := mngr.cb.Collection.Get("SCTEST:test_scorecard", nil)
	if err != nil {
		return nil, fmt.Errorf("mysql_test_director error getting SCTEST:test_scorecard %w", err)
	}
	// get the unmarshalled document (the Content) from the result
	var scorecardCB map[string]interface{}
	err = scorecardDataIn.Content(&scorecardCB)
	if err != nil {
		return nil, fmt.Errorf("mysql_test_director error getting SCTEST:test_scorecard Content %w", err)
	}
	return scorecardCB, nil
}

func upsertTestDoc(mngr *Manager, test_doc_file string, test_doc_id string) error {
	loadEnvironmentFile()
	// read the test document from the test file
	testScorcardFile := test_doc_file
	if _, err := os.Stat(testScorcardFile); err != nil {
		return fmt.Errorf("upsertTestDoc error reading test scorecard file %w", err)
	}
	scorecardBytes, _ := os.ReadFile(testScorcardFile)
	var scorecard map[string]interface{}
	err := json.Unmarshal(scorecardBytes, &scorecard)
	if err != nil {
		return fmt.Errorf("upsertTestDoc error unmarshalling test scorecard file %w", err)
	}
	// upsert the test scorecard document
	_, err = mngr.cb.Collection.Upsert(test_doc_id, scorecard, nil)
	if err != nil {
		return fmt.Errorf("upsertTestDoc error upserting test scorecard file %w", err)
	}
	return nil
}

func TestDirector_test_connection(t *testing.T) {
	defer goleak.VerifyNone(t)
	var cbCredentials director.DbCredentials
	var mysqlCredentials director.DbCredentials
	var err error
	loadEnvironmentFile()
	mysqlCredentials, cbCredentials, err = loadEnvironment()
	if err != nil {
		t.Fatal(fmt.Sprint("TestDirector_test_connection load environment error ", err))
	}
	if (director.DbCredentials{}) == cbCredentials {
		t.Errorf("loadEnvironment() error  did return cbCredentials from loadEnvironment")
		return
	}
	if (director.DbCredentials{} == mysqlCredentials) {
		t.Errorf("loadEnvironment() error  did return mysqlCredentials from loadEnvironment")
		return
	}
	documentID := "SCTEST:test_scorecard"
	t.Setenv("PROC_TESTING_ACCEPT_SCTEST_DOCIDS", "")
	mngr, _ := GetManager(documentID)
	err = getConnection(mngr, cbCredentials)
	if err != nil {
		t.Fatal(fmt.Sprint("TestDirector_test_connection Build GetConnection error ", err))
	}
	err = upsertTestDoc(mngr, "./testdata/test_scorecard.json", documentID)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error upserting test scorecard", err))
	}
	// get the test scorecard document (this is a Result - not a document)
	scorecardCB, err := getTestDoc(mngr)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error getting test scorecard from couchbase", err))
	}
	if scorecardCB == nil {
		t.Fatal("mysql_test_director error getting test scorecard from couchbase - scorecard is nil")
	}
}

func Test_loadEnvironment(t *testing.T) {
	defer goleak.VerifyNone(t)
	tests := []struct {
		name                 string
		wantMysqlCredentials director.DbCredentials
		wantCbCredentials    director.DbCredentials
		wantErr              bool
	}{
		{
			name:                 "test load environment",
			wantMysqlCredentials: director.DbCredentials{},
			wantCbCredentials:    director.DbCredentials{},
			wantErr:              false,
		},
	}
	loadEnvironmentFile()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMysqlCredentials, gotCbCredentials, err := loadEnvironment()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (director.DbCredentials{}) == gotCbCredentials {
				t.Errorf("loadEnvironment() error  did return CbCredentials from loadEnvironment")
				return
			}
			if (director.DbCredentials{} == gotMysqlCredentials) {
				t.Errorf("loadEnvironment() error  did return MysqlCredentials from loadEnvironment")
				return
			}
			if os.Getenv("CB_HOST") == "" {
				t.Errorf("loadEnvironment() error  did not find CB_HOST in environment")
				return
			}
			if os.Getenv("CB_USER") == "" {
				t.Errorf("loadEnvironment() error  did not find CB_HOST in environment")
				return
			}
			if os.Getenv("CB_PASSWORD") == "" {
				t.Errorf("loadEnvironment() error  did not find CB_USER in environment")
				return
			}
			if os.Getenv("CB_BUCKET") == "" {
				t.Errorf("loadEnvironment() error  did not find CB_BUCKET in environment")
				return
			}
			if os.Getenv("CB_SCOPE") == "" {
				t.Errorf("loadEnvironment() error  did not find CB_SCOPE in environment")
				return
			}
			if os.Getenv("CB_COLLECTION") == "" {
				t.Errorf("loadEnvironment() error  did not find CB_COLLECTION in environment")
				return
			}
			if os.Getenv("MYSQL_HOST") == "" {
				t.Errorf("loadEnvironment() error  did not find MYSQL_HOST in environment")
				return
			}
			if os.Getenv("MYSQL_USER") == "" {
				t.Errorf("loadEnvironment() error  did not find MYSQL_USER in environment")
				return
			}
			if os.Getenv("MYSQL_PASSWORD") == "" {
				t.Errorf("loadEnvironment() error  did not find MYSQL_PASSWORD in environment")
				return
			}
		})
	}
}

func Test_getQueryBlocks(t *testing.T) {
	defer goleak.VerifyNone(t)
	// setup a test document
	documentID := "SCTEST:test_scorecard"
	t.Setenv("PROC_TESTING_ACCEPT_SCTEST_DOCIDS", "")
	var mngr *Manager
	var err error
	loadEnvironmentFile()
	mngr, err = GetManager(documentID)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironment error GetManager %w", err))
	}
	var cbCredentials director.DbCredentials
	_, cbCredentials, err = loadEnvironment()
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironment error loadEnvironment %w", err))
	}
	err = getConnection(mngr, cbCredentials)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironment error getConnection %w", err))
	}
	err = upsertTestDoc(mngr, "./testdata/test_scorecard.json", documentID)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error upserting test scorecard", err))
	}

	// these two (results and queryMap return map[string]interface{})
	// (queryParams returns []interface{} - so it isn't here)
	tests := []struct {
		name    string
		args    *Manager
		want    []string
		wantErr bool
	}{
		{
			name:    "queryMaps",
			args:    mngr,
			want:    []string{"Block0", "Block1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		var retData map[string]interface{}
		var err error
		t.Run(tt.name, func(t *testing.T) {
			retData, err = getQueryBlocks(*tt.args)
			if retData == nil {
				t.Errorf("%v error = %v", tt.name, err)
			}
			got := director.ExtractKeys(retData)
			sort.Strings(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("getQueryBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getQueryBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSliceResultBlocks(t *testing.T) {
	defer goleak.VerifyNone(t)
	// setup a test document
	documentID := "SCTEST:test_scorecard"
	t.Setenv("PROC_TESTING_ACCEPT_SCTEST_DOCIDS", "")
	var mngr *Manager
	var err error
	loadEnvironmentFile()
	mngr, err = GetManager(documentID)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironment error GetManager %w", err))
	}
	var cbCredentials director.DbCredentials
	_, cbCredentials, err = loadEnvironment()
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironment error loadEnvironment %w", err))
	}
	err = getConnection(mngr, cbCredentials)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironment error getConnection %w", err))
	}
	err = upsertTestDoc(mngr, "./testdata/test_scorecard.json", documentID)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error upserting test scorecard", err))
	}

	// these two (results and queryMap return map[string]interface{})
	// (queryParams returns []interface{} - so it isn't here)
	tests := []struct {
		name    string
		args    *Manager
		want    []string
		wantErr bool
	}{
		{
			name:    "curves",
			args:    mngr,
			want:    []string{"application", "color", "control-data-source", "data-source", "forecast-length", "label", "level", "region", "statistic", "threshold", "truth", "valid-time", "variable"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		var retData []map[string]interface{}
		var err error
		t.Run(tt.name, func(t *testing.T) {
			retData, err = getPlotParamCurves(*tt.args)
			if retData == nil {
				t.Errorf("%v error = %v", tt.name, err)
			}
			got := director.ExtractKeys(retData[0])
			sort.Strings(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPlotParamCurves() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPlotParamCurves() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runManager(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Setenv("PROC_TESTING_ACCEPT_SCTEST_DOCIDS", "")
	var mngr *Manager
	var err error
	loadEnvironmentFile()
	tests := []struct {
		name            string
		docId           string
		fileName        string
		expectedSeconds int
	}{
		// {
		// 	name:            "test_1_Hour_Precipitation",
		// 	docId:           "SCTEST:test_1_Hour_Precipitation",
		// 	fileName:        "./testdata/test_1_Hour_Precipitation.json",
		// 	expectedSeconds: 60,
		// },
		{
			name:            "test_24_Hour_Precipitation",
			docId:           "SCTEST:test_24_Hour_Precipitation",
			fileName:        "./testdata/test_24_Hour_Precipitation.json",
			expectedSeconds: 5,
		},
		{
			name:            "test_90day_rufs_a_scorecard",
			docId:           "SCTEST:test_90day_rufs_a_scorecard",
			fileName:        "./testdata/test_90day_rufs_a_scorecard.json",
			expectedSeconds: 100,
		},
		{
			name:            "test_AMDAR",
			docId:           "SCTEST:test_AMDAR",
			fileName:        "./testdata/test_AMDAR.json",
			expectedSeconds: 60,
		},
		{
			name:            "test_Anomaly_Correlation",
			docId:           "SCTEST:test_Anomaly_Correlation",
			fileName:        "./testdata/test_Anomaly_Correlation.json",
			expectedSeconds: 5,
		},
		{
			name:            "test_Ceiling",
			docId:           "SCTEST:test_Ceiling",
			fileName:        "./testdata/test_Ceiling.json",
			expectedSeconds: 10,
		},
		{
			name:            "test_Composite_Reflectivity",
			docId:           "SCTEST:test_Composite_Reflectivity",
			fileName:        "./testdata/test_Composite_Reflectivity.json",
			expectedSeconds: 10,
		},
		{
			name:            "test_Echo_Top",
			docId:           "SCTEST:test_Echo_Top",
			fileName:        "./testdata/test_Echo_Top.json",
			expectedSeconds: 15,
		},
		{
			name:            "test_Gauge_Precipitation",
			docId:           "SCTEST:test_Gauge_Precipitation",
			fileName:        "./testdata/test_Gauge_Precipitation.json",
			expectedSeconds: 15,
		},
		{
			name:            "test_RAOBs_(GDAS)",
			docId:           "SCTEST:test_RAOBs_(GDAS)",
			fileName:        "./testdata/test_RAOBs_(GDAS).json",
			expectedSeconds: 20,
		},
		{
			name:            "test_RAOBs_(Traditional)",
			docId:           "SCTEST:test_RAOBs_(Traditional)",
			fileName:        "./testdata/test_RAOBs_(Traditional).json",
			expectedSeconds: 120,
		},
		{
			name:            "test_Sub_24_Hour_Precipitation",
			docId:           "SCTEST:test_Sub_24_Hour_Precipitation",
			fileName:        "./testdata/test_Sub_24_Hour_Precipitation.json",
			expectedSeconds: 5,
		},
		{
			name:            "test_Surface",
			docId:           "SCTEST:test_Surface",
			fileName:        "./testdata/test_Surface.json",
			expectedSeconds: 20,
		},
		{
			name:            "test_Surface_Land_Use",
			docId:           "SCTEST:test_Surface_Land_Use",
			fileName:        "./testdata/test_Surface_Land_Use.json",
			expectedSeconds: 40,
		},
		{
			name:            "test_Vertically_Integrated_Liquid",
			docId:           "SCTEST:test_Vertically_Integrated_Liquid",
			fileName:        "./testdata/test_Vertically_Integrated_Liquid.json",
			expectedSeconds: 15,
		},
		{
			name:            "test_Visibility",
			docId:           "SCTEST:test_Visibility",
			fileName:        "./testdata/test_Visibility.json",
			expectedSeconds: 10,
		},
		{
			name:            "test_flipped_scorecard",
			docId:           "SCTEST:test_flipped_scorecard",
			fileName:        "./testdata/test_flipped_scorecard.json",
			expectedSeconds: 60,
		},
	}
	var cbCredentials director.DbCredentials
	_, cbCredentials, err = loadEnvironment()
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironment error loadEnvironment %w", err))
	}

	for _, tt := range tests {
		log.Printf("Starting test %s", tt.name)
		start := time.Now()
		mngr, err = GetManager(tt.docId)
		if err != nil {
			t.Fatal(fmt.Errorf("manager - getManager for %s error  %w", tt.name, err))
		}
		err = getConnection(mngr, cbCredentials)
		if err != nil {
			t.Fatal(fmt.Errorf("manager loadEnvironmenttest %s error getConnection %w", tt.name, err))
		}
		err = upsertTestDoc(mngr, tt.fileName, tt.docId)
		if err != nil {
			t.Fatal(fmt.Errorf("manager upsertTestDoc test %s error upserting test scorecard %w", tt.name, err))
		}
		// get a manager
		manager, err := newScorecardManager(tt.docId)
		if err != nil {
			t.Fatal(fmt.Errorf("manager test %s NewScorecardManager error getting a manager %w", tt.name, err))
		}
		err = manager.Run()
		if err != nil {
			t.Fatal(fmt.Errorf("manager test %s Run error %w", tt.name, err))
		}
		elapsed := time.Since(start)
		if tt.expectedSeconds < int(elapsed.Seconds()) {
			t.Fatalf("manager test %s expected %d seconds but took %d seconds", tt.name, tt.expectedSeconds, int(elapsed.Seconds()))
		}
		log.Printf("The test %s took combined %s", tt.name, elapsed)
	}
}
