// +build integration

package mysql_director_test

import (
	"errors"
	"fmt"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/mysql_director"
	"github.com/couchbase/gocb/v2"
	"log"
	"os"
	"strconv"
	"testing"
)


func TestDirector_test_connection(t *testing.T) {
	var filename = os.Getenv("HOME") + strconv.QuoteRune(os.PathSeparator) + "adb-cb4-credentials"
	if ! mysql_director.checkFileExists(filename) {
		t.Fatal(fmt.Sprint("credential file does not exist :", filename))
	}
	var cb_connection *mysql_director.CB_connection = mysql_director.GetConnection()

	// read the test document from the test file
	var filename = "test_scorecard.json"
	if ! checkFileExists(filename) {
		t.Fatal(fmt.Sprintf("mysql_test_director error cannot open test scorecard document: ", filename," error: ", err))
	}
	scorecard, _ := ioutil.ReadFile(filename)
	var scorecardData interface{}
	err := json.Unmarshal(scorecardData, &data)
	if err != nil {
		t.Fatal(fmt.Sprintf("mysql_test_director error reading test scorecard", err))
	}
	// upsert the test scorecard document
	_, err = col.Upsert("u:jade",data, nil)
	if err != nil {
		t.Fatal(fmt.Sprintf("mysql_test_director error upserting test scorecard", err))
	}
	// get the test scorecard document
	var scorecardDataIn *gocb.GetResult
	var err error
	scorecardDataIn, err = cb_connection.CB_collection.Get("documentId", nil)
	if err != nil {
		t.Fatal(fmt.Sprintf("mysql_test_director error testing connection", err))
	}
	if ! reflect.DeepEqual(scoreData, scoredataIn) {
		t.Fatal(fmt.Sprintf("mysql_test_director test scorecard from file and retrieved scorecard from couchbase are not equal")
	}
}
