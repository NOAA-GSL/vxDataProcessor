package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/director"
	"github.com/couchbase/gocb/v2"
)

func TestDirector_test_connection(t *testing.T) {
	var cbCredentials director.DbCredentials
	var mysqlCredentials director.DbCredentials
	var err error
	mysqlCredentials, cbCredentials, err = loadEnvironmant(fmt.Sprint(os.Getenv("HOME"),"/vxDataProcessor.env"))
	if err != nil {
		t.Fatal(fmt.Sprint("TestDirector_test_connection load environment error ", err))
	}
	if (director.DbCredentials{}) == cbCredentials {
		t.Errorf("loadEnvironmant() error  did return cbCredentials from loadEnvironment")
		return
	}
	if  (director.DbCredentials{} == mysqlCredentials) {
		t.Errorf("loadEnvironmant() error  did return mysqlCredentials from loadEnvironment")
		return
	}
	cb_connection, err := getConnection(cbCredentials)
	if err != nil {
		t.Fatal(fmt.Sprint("TestDirector_test_connection Build GetConnection error ", err))
	}
	// read the test document from the test file
	testScorcardFile := "./testdata/test_scorecard.json"
	if _, err := os.Stat(testScorcardFile); err != nil {
		t.Fatal(fmt.Sprint("manager error reading test scorecard file", err))
	 }
	 var scorecardBytes, _ = os.ReadFile(testScorcardFile)
	var scorecard interface{}
	err = json.Unmarshal(scorecardBytes, &scorecard)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error unmarshalling test scorecard file", err))
	}
	// upsert the test scorecard document
	_, err = cb_connection.Collection.Upsert("MDTEST:test_scorecard", scorecard, nil)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error upserting test scorecard", err))
	}
	// get the test scorecard document (this is a Result - not a document)
	var scorecardDataIn *gocb.GetResult
	scorecardDataIn, err = cb_connection.Collection.Get("MDTEST:test_scorecard", nil)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error getting MDTEST:test_scorecard", err))
	}
	// get the unmarshalled document (the Content) from the result
	var scorecardCB interface{}
	err = scorecardDataIn.Content(&scorecardCB)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error getting MDTEST:test_scorecard Content", err))
	}
	// do a deep compare of the original and the retrieved unmarshalled document
	if !reflect.DeepEqual(scorecard, scorecardCB) {
		t.Fatal("mysql_test_director test scorecard from file and retrieved scorecard from couchbase are not equal")
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
			name: "test load environment",
			args: fmt.Sprint(os.Getenv("HOME"),"/vxDataProcessor.env"),
			wantMysqlCredentials: director.DbCredentials{},
			wantCbCredentials: director.DbCredentials{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMysqlCredentials, gotCbCredentials, err := loadEnvironmant(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadEnvironmant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (director.DbCredentials{}) == gotCbCredentials {
				t.Errorf("loadEnvironmant() error  did return CbCredentials from loadEnvironment")
				return
			}
			if  (director.DbCredentials{} == gotMysqlCredentials) {
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
