package mysql_director

import (
	"bufio"
	"errors"
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

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

func GetCredentials() (*CB_credentials, error) {
	var cb_credentials  = CB_credentials{}
	cb_credentials.cb_scope = "_default"
	var filename = os.Getenv("HOME") + strconv.QuoteRune(os.PathSeparator) + "adb-cb4-credentials"
	if ! checkFileExists(filename) {
		log.Print(fmt.Sprint("mysql_director  - credential does not exist - ", filename))
		return nil, errors.New(fmt.Sprint("mysql_director error credential file does not exist", filename))
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Print(fmt.Sprint("mysql_director - credential file open error - ", err))
		return nil, errors.New(fmt.Sprint("mysql_director error opening credential file", err))
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
			// for scorecards the collection is always 'SCORECARD'
			cb_credentials.cb_collection = "SCORECARD"
		default: // do nothing
		}
	}
	if cb_credentials.cb_host == "" {
        return nil, errors.New(fmt.Sprint("mysql_director cb_credentials.cb_host has not been set", filename))
    }
	if cb_credentials.cb_user == "" {
        return nil, errors.New(fmt.Sprint("mysql_director cb_credentials.cb_user has not been set", filename))
    }
	if cb_credentials.cb_password == "" {
        return nil, errors.New(fmt.Sprint("mysql_director cb_credentials.cb_password has not been set", filename))
    }
	if cb_credentials.cb_bucket == "" {
        return nil, errors.New(fmt.Sprint("mysql_director cb_credentials.cb_bucket has not been set", filename))
    }
	if cb_credentials.cb_collection == "" {
        return nil, errors.New(fmt.Sprint("mysql_director cb_credentials.cb_collection has not been set", filename))
    }
	return &cb_credentials, nil
}

func GetConnection() (*CB_connection, error) {
	var cb_credentials, err = GetCredentials()
	if err != nil {
        return nil, errors.New(fmt.Sprint("mysql_director GetCredentials error ", err))
    }

	var cb_connection CB_connection
	var options = gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cb_credentials.cb_user,
			Password: cb_credentials.cb_password,
		},
	}
	if err = options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment); err != nil {
        return nil, errors.New(fmt.Sprint("mysql_director gocb ApplyProfile error ", err))
	}
	// Initialize the Connection
	var cluster *gocb.Cluster
	cluster, err = gocb.Connect("couchbases://"+ cb_credentials.cb_host, options)
	if err != nil {
        return nil, errors.New(fmt.Sprint("mysql_director gocb Connect error ", err))
	}
	cb_connection.CB_cluster = cluster
	cb_connection.CB_bucket = cluster.Bucket(cb_credentials.cb_bucket)
	err = cb_connection.CB_bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
        return nil, errors.New(fmt.Sprint("mysql_director CB_bucket.WaitUntilReady error ", err))
	}
	cb_connection.CB_scope = cb_connection.CB_bucket.Scope(cb_credentials.cb_scope)
	col := cb_connection.CB_scope.Collection(cb_credentials.cb_collection)
	cb_connection.CB_collection = col
	return &cb_connection, nil
}

func Build(documentId string) error {
	// get the connection
	var cb_connection, err = GetConnection()
	if err != nil {
        return errors.New(fmt.Sprint("mysql_director Build GetConnection error ", err))
	}

	// get the scorecard document
	var docOut *gocb.GetResult
	docOut, err = cb_connection.CB_collection.Get(documentId, nil)
	if err != nil {
        return errors.New(fmt.Sprint("mysql_director Build GetResult error ", err))
	}
	fmt.Print(docOut)
	
	// build the input data elements and
	// for all the input elements fire off a thread to do the compute
	var cellPtr = builder.GetBuilder("TwoSampleTTest")
	err = cellPtr.SetGoodnessPolarity(gp)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director Build SetGoodnessPolarity error ", err))
	}
	err = cellPtr.SetMinorThreshold(minorThreshold)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetMinorThreshold - error message : ", err))
	}
	err = cellPtr.SetMajorThreshold(majorThreshold)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetMajorThreshold - error message : ", err))
	}
	err = cellPtr.SetInputData(inputData)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - SetInputData - error message : ", err))
	}
	err = cellPtr.ComputeSignificance(cellPtr.Data)
	if err != nil {
		return errors.New(fmt.Sprint("mysql_director - build - ComputeSignificance - error message : ", err))
	}
	// insert the elements into the in-memory document
	// upsert the document
	return nil
}
