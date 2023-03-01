package mysql_director

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"github.com/couchbase/gocb/v2"
)

type CB_credentials struct {
	cb_host string
	cb_user string
	cb_password string
	cb_bucket string
	cb_scope string
	cb_collection string
}

type CB_connection struct {
	CB_cluster *gocb.Cluster
	CB_bucket *gocb.Bucket
	CB_scope *gocb.Scope
	CB_collection *gocb.Collection
}

var gp = builder.GoodnessPolarity(1)
var minorThreshold = builder.Threshold(0.05)
var majorThreshold = builder.Threshold(0.01)
var inputData = builder.DerivedDataElement{
	CtlPop: []float64{0.2, 1.3, 3.2, 4.5},
	ExpPop: []float64{0.1, 1.5, 3.0, 4.1},
}

func TestString() string {
	return "this is a string from mysql_director"
}

func getCredentials() *CB_credentials {
	var cb_credentials  = CB_credentials{}
	cb_credentials.cb_scope = "_default"
	var filename = os.Getenv("HOME") + strconv.QuoteRune(os.PathSeparator) + "adb-cb4-credentials"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(fmt.Sprint("director - credential file open error - ", err))
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string
	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	// The method os.File.Close() is called
	// on the os.File object to close the file
	file.Close()
	// and then a loop iterates through
	// and prints each of the slice values.
	for _, each_ln := range text {
		s := strings.Split(each_ln, ":")
		switch s[0] {
		case "host":
			cb_credentials.cb_host = s[1]
		case "cb_user":
			cb_credentials.cb_user = s[1]
		case "cb_password":
			cb_credentials.cb_password = s[1]
		case "cb_bucket":
			cb_credentials.cb_bucket = s[1]
		case "cb_collection":
			cb_credentials.cb_collection = s[1]
		default: // do nothing
		}
	}
	return &cb_credentials
}

func GetConnection() *CB_connection {
	var cb_credentials *CB_credentials = getCredentials()
	var cb_connection CB_connection
	var options = gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cb_credentials.cb_user,
			Password: cb_credentials.cb_password,
		},
	}
	if err := options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment); err != nil {
		log.Fatal(err)
	}
	// Initialize the Connection
	cluster, err := gocb.Connect("couchbases://"+ cb_credentials.cb_host, options)
	if err != nil {
		log.Fatal(err)
	}
	cb_connection.CB_cluster = cluster
	cb_connection.CB_bucket = cluster.Bucket(cb_credentials.cb_bucket)
	err = cb_connection.CB_bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}
	cb_connection.CB_scope = cb_connection.CB_bucket.Scope(cb_credentials.cb_scope)
	col := cb_connection.CB_scope.Collection(cb_credentials.cb_collection)
	cb_connection.CB_collection = col
	return &cb_connection
}

func Build(documentId string) {
	// get the connection
	var cb_connection = GetConnection()
	// get the scorecard document
	var docOut *gocb.GetResult
	var err error
	docOut, err = cb_connection.CB_collection.Get(documentId, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(docOut)
	// builde the input data elements and
	// for all the input elements fire off a thread to do the compute
	var cellPtr = builder.GetBuilder("TwoSampleTTest")
	err = cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetGoodnessPolarity - error message : ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(inputData)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		log.Fatal(fmt.Sprint("director - build - ComputeSignificance - error message : ", err))
	}
	// insert the elements into the in-memory document
	// upsert the document
}
