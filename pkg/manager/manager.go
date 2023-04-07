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
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/client"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/director"
	"github.com/couchbase/gocb/v2"
)

func loadEnvironmant() (mysqlCredentials, cbCredentials director.DbCredentials, err error) {
	cbCredentials = director.DbCredentials{
		Scope:      "_default",
		Collection: "SCORECARD",
		Bucket:     os.Getenv("CB_BUCKET"),
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
func getConnection(mngr *Manager, cbCredentials director.DbCredentials) (err error) {
	options := gocb.ClusterOptions{
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

func upsertSubDocument(mngr Manager, path string, subDoc interface{}) error {
	mops := []gocb.MutateInSpec{
		gocb.UpsertSpec(path, subDoc, &gocb.UpsertSpecOptions{}),
	}
	upsertResult, err := mngr.cb.Collection.MutateIn(mngr.documentId, mops, &gocb.MutateInOptions{
		Timeout: 10050 * time.Millisecond,
	})
	if err != nil {
		return fmt.Errorf("manager upsertSubDocument error: %q", err)
	}
	// There is probably a better way to do this
	if upsertResult.MutationToken().BucketName() != "vxdata" {
		return fmt.Errorf("manager upsertSubDocument result bad upsertResult")
	}
	return nil
}

func getSubDocument(mngr Manager, path string, subDocPtr *interface{}) error {
	ops := []gocb.LookupInSpec{
		gocb.GetSpec(path, &gocb.GetSpecOptions{IsXattr: false}),
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

// retrieve the queryMap.blocks section of the document by subdoc get
func getQueryBlocks(mngr Manager) (map[string]interface{}, error) {
	var blocks interface{}
	err := getSubDocument(mngr, "queryMap.blocks", &blocks)
	if err != nil {
		return nil, fmt.Errorf("manager getQueryBlocks error %q", err)
	}
	return blocks.(map[string]interface{}), err
}

// retrieve the PlotParams section of the document by subdoc get
func getPlotParams(mngr Manager) (map[string]interface{}, error) {
	var plotParams interface{}
	err := getSubDocument(mngr, "plotParams", &plotParams)
	if err != nil {
		return nil, fmt.Errorf("manager getPlotParams error %q", err)
	}
	return plotParams.(map[string]interface{}), err
}

// retrieve the PlotParam.curves (this is an array) section of the document by subdoc get
func getPlotParamCurves(mngr Manager) ([]map[string]interface{}, error) {
	var curves interface{}
	var curveArray []map[string]interface{}
	err := getSubDocument(mngr, "plotParams.curves", &curves)
	if err != nil {
		return nil, fmt.Errorf("manager getPlotParamCurves error %q", err)
	}
	for _, c := range curves.([]interface{}) {
		curveArray = append(curveArray, c.(map[string]interface{}))
	}
	return curveArray, err
}

// retrieve the dateRange section of the document by subdoc get
// and convert it to a dateRange struct
func getDateRange(mngr Manager) (director.DateRange, error) {
	var datesStr interface{}
	err := getSubDocument(mngr, "dateRange", &datesStr)
	var dateRange director.DateRange
	// parse the daterange string
	// "02/19/2023 20:00 - 03/21/2023 20:00"
	if err != nil {
		return dateRange, fmt.Errorf("manager getDateRange error %q", err)
	}
	dateParts := strings.Split(datesStr.(string), " - ")
	fromTime, err := time.Parse("01/02/2006 15:04", dateParts[0])
	if err != nil {
		return dateRange, fmt.Errorf("manager getDataRange error converting from date to epoch error %q", err)
	}
	fromSecs := fromTime.Unix()
	if err != nil {
		return dateRange, fmt.Errorf("manager getDataRange error getting date range from document error %q", err)
	}
	toTime, err := time.Parse("01/02/2006 15:04", dateParts[1])
	if err != nil {
		return dateRange, fmt.Errorf("manager getDataRange error converting from date to epoch error %q", err)
	}
	toSecs := toTime.Unix()
	dateRange.FromSecs = fromSecs
	dateRange.ToSecs = toSecs
	return dateRange, err
}

func convertStdToPercent(std string) (percent float64, err error) {
	stdfloat, err := strconv.ParseFloat(std, 64)
	if err != nil {
		err = fmt.Errorf("manager convertStdToPercent error converting standard deviation %q to percent error: %q", std, err)
		return 0, err
	}
	// round to nearest int - should be 1, 2, or 3 - fractions are not allowed
	stdint := int(stdfloat + 0.5)
	switch stdint {
	case 1:
		percent = 68
	case 2:
		percent = 95
	case 3:
		percent = 99.7
	default:
		err = fmt.Errorf("manager convertStdToPercent error converting standard deviation %q - not between 1 and 3 inclusive", std)
		return 0, err
	}
	return percent, err
}

func getThresholds(plotParams map[string]interface{}) (minorThreshold float64, majorThreshold float64, err error) {
	percentStddev := plotParams["scorecard-percent-stdv"]
	switch percentStddev {
	case "Percent":
		minorThreshold, err = strconv.ParseFloat(plotParams["minor-threshold-by-percent"].(string), 64)
		if err != nil {
			return minorThreshold, majorThreshold, err
		}
		majorThreshold, err = strconv.ParseFloat(plotParams["major-threshold-by-percent"].(string), 64)
		if err != nil {
			return minorThreshold, majorThreshold, err
		}
	case "Standard Deviation":
		minorThreshold, err = convertStdToPercent(plotParams["minor-threshold-by-stdv"].(string))
		if err != nil {
			return minorThreshold, majorThreshold, err
		}
		majorThreshold, err = convertStdToPercent(plotParams["major-threshold-by-stdv"].(string))
		if err != nil {
			return minorThreshold, majorThreshold, err
		}
	default:
		return minorThreshold, majorThreshold, fmt.Errorf("manager Run error getting threshold percentages")
	}
	return minorThreshold, majorThreshold, nil
}

func notifyMatsRefresh(scorecardAppUrl string, docId string) error {
	err := client.NotifyScorecard(scorecardAppUrl, docId)
	if err != nil {
		return fmt.Errorf("manager notifyMATSRefresh error: %v", err)
	}
	return err
}

func processRegion(
	mngr Manager,
	appName string,
	queryRegionName string,
	queryRegion map[string]interface{},
	blockRegionName string,
	region *interface{},
	regionPath string,
	mysqlCredentials director.DbCredentials,
	dateRange director.DateRange,
	minorThreshold float64,
	majorThreshold float64,
	documentScorecardAppUrl string,
) error {
	if strings.ToUpper(appName) == "CB" {
		log.Print("launch CB director - which we don't have yet")
	} else {
		// launch mysql director
		mysqlDirector, err := director.GetDirector("MysqlDirector", mysqlCredentials, dateRange, minorThreshold, majorThreshold)
		if err != nil {
			err = fmt.Errorf("manager Run error getting director: %q", err)
			return err
		}
		*region, err = mysqlDirector.Run(*region, queryRegion)
		if err != nil {
			err = fmt.Errorf("manager Run error running director: %q", err)
			return err
		}
	}
	err := upsertSubDocument(mngr, regionPath, region)
	if err != nil {
		return fmt.Errorf("manager Run error upserting resultRegion: %q error: %q", blockRegionName, err)
	}
	// notify server to update with scorecardApUrl
	// try to get the SCORECARD_APP_URL from the environment
	scorecardAppUrl := os.Getenv("DEBUG_SCORECARD_APP_URL")
	if scorecardAppUrl == "" {
		// not in environment. so use the one from the document
		scorecardAppUrl = documentScorecardAppUrl
	}
	err = notifyMatsRefresh(scorecardAppUrl, mngr.documentId)
	if err != nil {
		return fmt.Errorf("manager Run error Failed to Notify appUrl %q: error: %q", scorecardAppUrl, err)
	}
	return nil
}

func (mngr Manager) Run() (err error) {
	// load the environment
	var mysqlCredentials, cbCredentials director.DbCredentials
	var minorThreshold float64
	var majorThreshold float64
	// initially unknown
	mysqlCredentials, cbCredentials, err = loadEnvironmant()
	if err != nil {
		return fmt.Errorf("manager loadEnvironmant error %q", err)
	}
	err = getConnection(&mngr, cbCredentials)
	if err != nil {
		return fmt.Errorf("manager Run GetConnection error: %q", err)
	}
	queryBlocks, err := getQueryBlocks(mngr)
	if err != nil {
		err = fmt.Errorf("manager Run error getting queryBlocks: %q", err)
		return err
	}
	plotParams, err := getPlotParams(mngr)
	if err != nil {
		err = fmt.Errorf("manager Run error getting plotParamCurves: %q", err)
		return err
	}
	minorThreshold, majorThreshold, err = getThresholds(plotParams)
	if err != nil {
		err = fmt.Errorf("manager Run error getting thresholds: %q", err)
		return err
	}
	curves, err := getPlotParamCurves(mngr)
	if err != nil {
		err = fmt.Errorf("manager Run error getting plotParamCurves: %q", err)
		return err
	}
	dateRange, err := getDateRange(mngr)
	if err != nil {
		err = fmt.Errorf("manager Run error getting daterange: %q", err)
		return err
	}
	numCurves := len(curves)
	// blocks and queryBlocks have the same keys
	blockKeys := director.Keys(queryBlocks)
	sort.Strings(blockKeys)
	numBlocks := len(blockKeys)
	for i := 0; i < numBlocks; i++ {
		blockName := blockKeys[i]
		var block interface{}
		err = getSubDocument(mngr, "results.blocks."+blockName, &block)
		if err != nil {
			return fmt.Errorf("manager Run error getting block result %q", err)
		}
		scorecardAppUrl := block.(map[string]interface{})["blockApplication"].(string)
		queryBlock := queryBlocks[blockKeys[i]].(map[string]interface{})
		var appName string
		for i := 0; i < numCurves; i++ {
			curve := curves[i]
			if curve["label"] == block.(map[string]interface{})["blockTitle"].(map[string]interface{})["label"] {
				appName = curve["application"].(string)
				break
			}
		}
		queryData := queryBlock["data"].(map[string]interface{})
		blockRegionNames := director.Keys(block.(map[string]interface{})["data"].(map[string]interface{}))
		sort.Strings(blockRegionNames)
		queryRegionNames := director.Keys(queryData)
		sort.Strings(queryRegionNames)
		numBlockRegions := len(blockRegionNames)
		numQueryRegions := len(queryRegionNames)
		if numBlockRegions != numQueryRegions {
			return fmt.Errorf("manager Run Number of block regions %v does not equal the number of query regions %v", numBlockRegions, numQueryRegions)
		}
		if !reflect.DeepEqual(blockRegionNames, queryRegionNames) {
			return fmt.Errorf("manager block regions list %v does not equal query regions list %v", blockRegionNames, queryRegionNames)
		}
		for i := 0; i < numBlockRegions; i++ {
			queryRegionName := queryRegionNames[i]
			queryRegion := queryData[queryRegionName].(map[string]interface{})
			blockRegionName := blockRegionNames[i]
			var region interface{}
			regionPath := "results.blocks." + blockName + ".data." + blockRegionName
			err = getSubDocument(mngr, regionPath, &region)
			if err != nil {
				return fmt.Errorf("error getting region SubDocument %q", err)
			}
			err = processRegion(mngr,
				appName,
				queryRegionName,
				queryRegion,
				blockRegionName,
				&region,
				regionPath,
				mysqlCredentials,
				dateRange,
				minorThreshold,
				majorThreshold,
				scorecardAppUrl)
			if err != nil {
				return fmt.Errorf("error processing scorecard Run %q", err)
			}
		}
	}
	return nil
}

var myScorecardManager = Manager{}

func newScorecardManager(documentId string) (*Manager, error) {
	myScorecardManager.cb = &cbConnection{}
	myScorecardManager.documentId = documentId
	return &myScorecardManager, nil
}
