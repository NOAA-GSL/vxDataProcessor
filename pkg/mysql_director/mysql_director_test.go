package mysql_director_test

import (
	"fmt"
	"log"
	"testing"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/mysql_director"
	"github.com/couchbase/gocb/v2"
)

func TestDirector_test_connection(t *testing.T) {
	// var cb_connection *mysql_director.CB_connection = mysql_director.GetConnection()
	// //cb_connection.cluster QUERY!!
	// 	// get the scorecard document
	// 	var docOut *gocb.GetResult
	// 	var err error
	// 	docOut, err = cb_connection.CB_collection.Get("documentId", nil)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(docOut)
		if mysql_director.TestString() != "this is a string from mysql_director" {
			t.Fatal("Wrong test string :")
		}
}