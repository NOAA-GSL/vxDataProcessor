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
		return nil, fmt.Errorf("mysql_test_director error getting SCTEST:test_scorecard %q", err)
	}
	// get the unmarshalled document (the Content) from the result
	var scorecardCB map[string]interface{}
	err = scorecardDataIn.Content(&scorecardCB)
	if err != nil {
		return nil, fmt.Errorf("mysql_test_director error getting SCTEST:test_scorecard Content %v", err)
	}
	return scorecardCB, nil
}

func upsertTestDoc(mngr *Manager) error {
	loadEnvironmentFile()
	// read the test document from the test file
	testScorcardFile := "./testdata/test_scorecard.json"
	if _, err := os.Stat(testScorcardFile); err != nil {
		return fmt.Errorf("upsertTestDoc error reading test scorecard file %v", err)
	}
	scorecardBytes, _ := os.ReadFile(testScorcardFile)
	var scorecard map[string]interface{}
	err := json.Unmarshal(scorecardBytes, &scorecard)
	if err != nil {
		return fmt.Errorf("upsertTestDoc error unmarshalling test scorecard file %v", err)
	}
	// upsert the test scorecard document
	_, err = mngr.cb.Collection.Upsert("SCTEST:test_scorecard", scorecard, nil)
	if err != nil {
		return fmt.Errorf("upsertTestDoc error upserting test scorecard file %v", err)
	}
	return nil
}

func TestDirector_test_connection(t *testing.T) {
	var cbCredentials director.DbCredentials
	var mysqlCredentials director.DbCredentials
	var err error
	loadEnvironmentFile()
	mysqlCredentials, cbCredentials, err = loadEnvironmant()
	if err != nil {
		t.Fatal(fmt.Sprint("TestDirector_test_connection load environment error ", err))
	}
	if (director.DbCredentials{}) == cbCredentials {
		t.Errorf("loadEnvironmant() error  did return cbCredentials from loadEnvironment")
		return
	}
	if (director.DbCredentials{} == mysqlCredentials) {
		t.Errorf("loadEnvironmant() error  did return mysqlCredentials from loadEnvironment")
		return
	}
	var documentId string = "SCTEST:test_scorecard"
	mngr, _ := GetManager("SC", documentId)
	err = getConnection(mngr, cbCredentials)
	if err != nil {
		t.Fatal(fmt.Sprint("TestDirector_test_connection Build GetConnection error ", err))
	}
	err = upsertTestDoc(mngr)
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

func Test_loadEnvironmant(t *testing.T) {
	tests := []struct {
		name                 string
		args                 string
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
			gotMysqlCredentials, gotCbCredentials, err := loadEnvironmant()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadEnvironmant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (director.DbCredentials{}) == gotCbCredentials {
				t.Errorf("loadEnvironmant() error  did return CbCredentials from loadEnvironment")
				return
			}
			if (director.DbCredentials{} == gotMysqlCredentials) {
				t.Errorf("loadEnvironmant() error  did return MysqlCredentials from loadEnvironment")
				return
			}
			if os.Getenv("CB_HOST") == "" {
				t.Errorf("loadEnvironmant() error  did not find CB_HOST in environment")
				return
			}
			if os.Getenv("CB_USER") == "" {
				t.Errorf("loadEnvironmant() error  did not find CB_HOST in environment")
				return
			}
			if os.Getenv("CB_PASSWORD") == "" {
				t.Errorf("loadEnvironmant() error  did not find CB_USER in environment")
				return
			}
			if os.Getenv("CB_BUCKET") == "" {
				t.Errorf("loadEnvironmant() error  did not find CB_BUCKET in environment")
				return
			}
			if os.Getenv("CB_SCOPE") == "" {
				t.Errorf("loadEnvironmant() error  did not find CB_SCOPE in environment")
				return
			}
			if os.Getenv("CB_COLLECTION") == "" {
				t.Errorf("loadEnvironmant() error  did not find CB_COLLECTION in environment")
				return
			}
			if os.Getenv("MYSQL_HOST") == "" {
				t.Errorf("loadEnvironmant() error  did not find MYSQL_HOST in environment")
				return
			}
			if os.Getenv("MYSQL_USER") == "" {
				t.Errorf("loadEnvironmant() error  did not find MYSQL_USER in environment")
				return
			}
			if os.Getenv("MYSQL_PASSWORD") == "" {
				t.Errorf("loadEnvironmant() error  did not find MYSQL_PASSWORD in environment")
				return
			}
		})
	}
}

func Test_getQueryBlocks(t *testing.T) {
	// setup a test document
	var documentId string = "SCTEST:test_scorecard"
	var mngr *Manager
	var err error
	loadEnvironmentFile()
	mngr, err = GetManager("SC", documentId)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error GetManager %q", err))
	}
	var cbCredentials director.DbCredentials
	_, cbCredentials, err = loadEnvironmant()
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error loadEnvironmant %q", err))
	}
	err = getConnection(mngr, cbCredentials)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error getConnection %q", err))
	}
	err = upsertTestDoc(mngr)
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
			got := director.Keys(retData)
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
	// setup a test document
	var documentId string = "SCTEST:test_scorecard"
	var mngr *Manager
	var err error
	loadEnvironmentFile()
	mngr, err = GetManager("SC", documentId)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error GetManager %q", err))
	}
	var cbCredentials director.DbCredentials
	_, cbCredentials, err = loadEnvironmant()
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error loadEnvironmant %q", err))
	}
	err = getConnection(mngr, cbCredentials)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error getConnection %q", err))
	}
	err = upsertTestDoc(mngr)
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
			got := director.Keys(retData[0])
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
	// setup a test document
	var documentId string = "SCTEST:test_scorecard"
	var mngr *Manager
	var err error
	start := time.Now()
	loadEnvironmentFile()
	mngr, err = GetManager("SC", documentId)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error GetManager %q", err))
	}
	var cbCredentials director.DbCredentials
	_, cbCredentials, err = loadEnvironmant()
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error loadEnvironmant %q", err))
	}
	err = getConnection(mngr, cbCredentials)
	if err != nil {
		t.Fatal(fmt.Errorf("manager loadEnvironmant error getConnection %q", err))
	}
	err = upsertTestDoc(mngr)
	if err != nil {
		t.Fatal(fmt.Sprint("manager upsertTestDoc error upserting test scorecard", err))
	}
	// get a manager
	manager, err := NewScorecardManager(documentId)
	if err != nil {
		t.Fatal(fmt.Sprint("manager test NewScorecardManager error getting a manager", err))
	}
	err = manager.Run()
	if err != nil {
		t.Fatal(fmt.Sprint("manager test run error ", err))
	}
	elapsed := time.Since(start)
	fmt.Printf("The test took combined %s", elapsed)
}
