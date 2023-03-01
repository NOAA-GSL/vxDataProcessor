package director_test

import (
	"fmt"
	"log"
	"testing"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/director"
	"github.com/couchbase/gocb/v2"
)

func TestDirector_test_connection(t *testing.T) {
	var cb_connection director.CB_connection = director.GetConnection()
	//cb_connection.cluster QUERY!!
		// get the scorecard document
		var docOut *gocb.GetResult
		var err error
		docOut, err = cb_connection.CB_collection.Get("documentId", nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(docOut)
}