package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/manager"
	"github.com/couchbase/gocb/v2"
)

func TestDirector_test_connection(t *testing.T) {
	var filename = fmt.Sprint(os.Getenv("HOME"), "/adb-cb4-credentials")
	if !manager.CheckFileExists(filename) {
		t.Fatal(fmt.Sprint("credential file does not exist :", filename))
	}
	var cb_connection, err = manager.GetConnection()
	if err != nil {
		t.Fatal(fmt.Sprint("TestDirector_test_connection Build GetConnection error ", err))
	}

	// read the test document from the test file
	filename = "./testdata/test_scorecard.json"
	if !manager.CheckFileExists(filename) {
		t.Fatal(fmt.Sprint("mysql_test_director error cannot open test scorecard document: ", filename, " error: ", err))
	}
	var scorecardBytes, _ = os.ReadFile(filename)
	var scorecard interface{}
	err = json.Unmarshal(scorecardBytes, &scorecard)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error reading test scorecard", err))
	}
	// upsert the test scorecard document
	_, err = cb_connection.CB_collection.Upsert("MDTEST:test_scorecard", scorecard, nil)
	if err != nil {
		t.Fatal(fmt.Sprint("mysql_test_director error upserting test scorecard", err))
	}
	// get the test scorecard document (this is a Result - not a document)
	var scorecardDataIn *gocb.GetResult
	scorecardDataIn, err = cb_connection.CB_collection.Get("MDTEST:test_scorecard", nil)
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
