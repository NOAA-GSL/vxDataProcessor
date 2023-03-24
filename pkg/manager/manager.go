package manager

/*
The Manager has the following responsibilities and transformations.

1. The manager will maintain a Couchbase connection.
1. The manager is given a process_id from the service. The service
will have as many managers open as go workers as needed so that it can handle multiple service
requests simultaneously. The service starts a manager in a GO worker routine and
the manager is passed the id of the corresponding scorecard document.
1. The manager will read the scorcard document associated with the id from Couchbase
and maintain it in memory on behalf of its directors.
1. The manager will start go workers (which are directors) making sure that the number of
workers (directors) does not exceed the maximum number of database connections
configured for each kind of director. For example currently most apps are legacy apps
that require a mysql database connection. If the configuration specifies 20 allowed mysql
database connections the manager will allow up to twenty workers. Each worker is a
director and each director will maintain its own database connection (e.g. mysql client).
1. The appname associated with a scorecard block tells the manager what kind of director is
needed for each scorecard block. Each block requires an associated database query template
which is included in the scorecard document. The manager will build a queue of sc_element
structures each of which has a pointer to the associated scorecard section
(which has the query, the template variables e.g. region, statistic, variable,
they are the keys to the specific row). For example
... ```results..["rows"]["Row0"]["data"]["All HRRR domain"]["Bias (Model - Obs)"][]"2m RH"][.... ]```
1. Each director must derive a query (making appropriate substitutions to the template) for each
cell that needs to be calculated, then query the database for the cell data, format the data
into an InputDataElement and send the data element to an appropriate builder in a go routine. The
director uses as many GO routines as necessary to derive all the cells required of it. For example,
maybe this is one director per row, and the builder parts are delineated by region and forecastlen.
1. The builder will process the data for a given cell by...
   1. Matching the data by time.
   2. Processing the data for the associated statistic (like RMSE or BIAS).
   3. Processing the pvalue statistic.
   4. Return the result to the director.
1. The builders update the in-memory scorecard directly. When enough builders finish
the director will notify the manager when a scorecard upsert is necessary.
(perhaps when each row is complete, i.e. the director dies?)
1. The manager upserts the scorecard document with the current new results. There may be many of these upserts.
1. The manager knows that the results have all been processed when the directors have all died. The
manager does a final upsert of the scorecard, provides the return status for the service call
and then it politely dies.
*/

import (
	"fmt"
	"reflect"
	"os"
	"log"
	"time"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/director"
	"github.com/couchbase/gocb/v2"
	"github.com/joho/godotenv"
)

func loadEnvironmant(environmentFile string) (mysqlCredentials, cbCredentials director.DbCredentials, err error) {
	err = godotenv.Load(environmentFile)
	if err != nil {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Error loading .env file: %q", environmentFile)
	}

	cbCredentials = director.DbCredentials{
		Scope:      "_default",
		Collection: "SCORECARD",
		Bucket:		os.Getenv("CB_BUCKET"),
		Host:       os.Getenv("CB_HOST"),
	}

	if cbCredentials.Host == "" {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Undefined CB_HOST in environment")
	}
	cbCredentials.User = os.Getenv("CB_USER")
	if cbCredentials.User == "" {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Undefined CB_USER in environment")
	}
	cbCredentials.Password = os.Getenv("CB_PASSWORD")
	if cbCredentials.Password == "" {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Undefined CB_PASSWORD in environment")
	}
	cbCredentials.Bucket = os.Getenv("CB_BUCKET")
	if cbCredentials.Bucket == "" {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Undefined CB_BUCKET in environment")
	}

	// refer to https://github.com/go-sql-driver/mysql/#dsn-data-source-name
	mysqlCredentials.Host = os.Getenv("MYSQL_HOST")
	if mysqlCredentials.Host == "" {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Undefined MYSQL_HOST in environment")
	}
	mysqlCredentials.User = os.Getenv("MYSQL_USER")
	if mysqlCredentials.User == "" {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Undefined MYSQL_USER in environment")
	}
	mysqlCredentials.Password = os.Getenv("MYSQL_PASSWORD")
	if mysqlCredentials.Password == "" {
		return director.DbCredentials{}, director.DbCredentials{}, fmt.Errorf("Undefined MYSQL_PASSWORD in environment")
	}
	return mysqlCredentials, cbCredentials, nil
}

// get the couchbase connection
// mysql connections are maintained in the mysql_director
func getConnection(mngr *Manager,cbCredentials director.DbCredentials) (err error) {
	var options = gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cbCredentials.User,
			Password: cbCredentials.Password,
		},
	}
	if err = options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment); err != nil {
		return fmt.Errorf("manager gocb ApplyProfile error: %q", err)
	}
	// Initialize the Connection
	var cluster *gocb.Cluster
	cluster, err = gocb.Connect("couchbase://"+cbCredentials.Host, options)
	if err != nil {
		return fmt.Errorf("manager gocb Connect error: %q", err)
	}
 	mngr.cb.Cluster = cluster
	mngr.cb.Bucket = cluster.Bucket(cbCredentials.Bucket)
	err = mngr.cb.Bucket.WaitUntilReady(50*time.Second, nil)
	if err != nil {
		return fmt.Errorf("manager bucket.WaitUntilReady error: %q", err)
	}
	mngr.cb.Scope = mngr.cb.Bucket.Scope(cbCredentials.Scope)
	mngr.cb.Collection = mngr.cb.Bucket.Collection(cbCredentials.Collection)
	return nil
}

// return tags for the struct as a slice
func fields(s *struct{}) []string {
	var tags []string
	val := reflect.ValueOf(s)
    for i := 0; i < val.Type().NumField(); i++ {
        tags = append(tags, fmt.Sprint(val.Type().Field(i).Name))
    }
	return tags
}

func getMapSubDocument(mngr Manager, path string) (map[string]interface{}, error){
	ops := []gocb.LookupInSpec{
		gocb.GetSpec(path, &gocb.GetSpecOptions{IsXattr:false,}),
	}
	getResult, err := mngr.cb.Collection.LookupIn(mngr.documentId, ops, &gocb.LookupInOptions{})
	if err != nil {
		return nil, fmt.Errorf("manager getSubDocument LookupIn error %q", err)
	}
	var subDoc map[string]interface{}
	err = getResult.ContentAt(0, &subDoc)
	if err != nil {
		return nil, fmt.Errorf("manager getSubDocument getResult error %q", err)
	}
	return subDoc, nil
}

func getSubDocument(mngr Manager, path string, subDocPtr interface{}) error {
	ops := []gocb.LookupInSpec{
		gocb.GetSpec(path, &gocb.GetSpecOptions{IsXattr:false,}),
	}
	getResult, err := mngr.cb.Collection.LookupIn(mngr.documentId, ops, &gocb.LookupInOptions{})
	if err != nil {
		return fmt.Errorf("manager getSubDocument LookupIn error %q", err)
	}
	err = getResult.ContentAt(0, subDocPtr)
	if err != nil {
		return fmt.Errorf("manager getSubDocument getResult error %q", err)
	}
	return nil
}

// retrieve the results.blocks section of the document by subdoc get
func getResultBlocks(mngr Manager)(map[string]interface{}, error){
	var blocks map[string]interface{}
	err := getSubDocument(mngr, "SCORECARD.results.blocks", &blocks)
	if err != nil {
		return nil, fmt.Errorf("manager getResultBlocks error %q", err)
	}
	return blocks, err
}

// retrieve the queryMap.blocks section of the document by subdoc get
func getQueryBlocks(mngr Manager)(map[string]interface{}, error){
	var blocks map[string]interface{}
	err := getSubDocument(mngr, "SCORECARD.queryMap.blocks", &blocks)
	if err != nil {
		return nil, fmt.Errorf("manager getQueryBlocks error %q", err)
	}
	return blocks, err
}

// retrieve the PlotParam.curves (this is an array) section of the document by subdoc get
func getPlotParamCurves(mngr Manager)([]map[string]interface{}, error){
	var curves []map[string]interface{}
	err := getSubDocument(mngr, "SCORECARD.plotParams.curves", &curves)
	if err != nil {
		return nil, fmt.Errorf("manager getPlotParamCurves error %q", err)
	}
	return curves, err
}

func  (mngr Manager)Run() error {
	var err error
	// load the environment
	var mysqlCredentials, cbCredentials director.DbCredentials
	mysqlCredentials, cbCredentials, err = loadEnvironmant(mngr.environmentFile)
	if err != nil {
		return fmt.Errorf("manager loadEnvironmant error %q", err)
	}
	log.Printf("credentials %v",mysqlCredentials)
	err = getConnection(&mngr, cbCredentials)
	if err != nil {
		return fmt.Errorf("manager Build GetConnection error: %q", err)
	}
	// get the scorecard document
	// var scorecardDataIn *gocb.GetResult
	// scorecardDataIn, err = mngr.cb.Collection.Get(mngr.documentId, nil)
	// if err != nil {
	// 	return fmt.Errorf("manager Build error getting scorecard: %q  error: %q", mngr.documentId, err)
	// }
	// // get the unmarshalled document (the Content) from the result
	// var scorecard map[string]interface{}
	// err = scorecardDataIn.Content(scorecard)
	// if err != nil {
	// 	return fmt.Errorf("manager Build error getting scorecard Content: %q", err)
	// }
	// mngr.ScorecardCB = scorecard
	// // get the scorecardAppUrl so that manager can use it to notify
	// // the scorecard app to refresh its mongo data after the upsert
	// //scorecardApUrl := block["blockApplication"]
	// // iterate the rows in the scorecard
	blocks, err := getResultBlocks(mngr)
	if err != nil {
		err = fmt.Errorf("manager Build error getting resultBlocks: %q", err)
		return err
	}
	log.Printf("blocks: %v", blocks)
	queryBlocks, err := getQueryBlocks(mngr)
	if err != nil {
		err = fmt.Errorf("manager Build error getting queryBlocks: %q", err)
		return err
	}
	log.Printf("queryBlocks %v",queryBlocks)
	curves, err := getPlotParamCurves(mngr)
	if err != nil {
		err = fmt.Errorf("manager Build error getting plotParamCurves: %q", err)
		return err
	}
	log.Printf("curves%v",curves)

	// numCurves := curves.NumField()
    // for i := 0; i < numBlocks; i++ {
	// 	block := blocks.Field(i)
	// 	var appName string
	// 	for i := 0; i < numCurves; i++ {
	// 		curve := curves.Field(i)
	// 		if curve.Tag.Get("Label") == block.Tag.Get("BlockTitle").Tag.Get("Label"){
	// 			appName = curve.Application
	// 		}
	// 	}
	// 	data := block.Tag.Get("Data")
	// 	queryData := queryBlocks
	// 	numRegions := queryData.NumField()
	// 	for i := 0; i < numRegions; i++ {
	// 		region := queryData.Field(i).(struct{})
	// 		// launch a director for this region
	// 		if strings.ToUpper(appName) == "CB" {
	// 			// launch CB director - which we don't have yet
	// 		} else {
	// 			//launch mysql director
	// 			var mysqlDirector *director.Director
	// 			mysqlDirector, err := director.GetDirector("mySqlDirector", mysqlCredentials)
	// 			if err != nil {
	// 				err = fmt.Errorf("manager Build error getting director: %q", err)
	// 				return err
	// 			}
	// 			mysqlDirector.Run(regionMap, queryRegionMap)
	// 		}
	// 	}
	//}
		// upsert the document
		return nil
	}

var myScorecardManager = Manager{}
func NewScorecardManager(environmentFile, documentId string) (*Manager, error) {
	myScorecardManager.environmentFile = environmentFile
	myScorecardManager.ScorecardCB = nil
	myScorecardManager.cb = &cbConnection{}
	myScorecardManager.documentId = documentId
	return &myScorecardManager, nil
}
