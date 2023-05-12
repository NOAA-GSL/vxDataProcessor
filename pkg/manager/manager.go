package manager

/*
The Manager has the following responsibilities and transformations.

1. The manager will maintain a Couchbase connection.
1. The manager is given a documentID from the API service. The API service
will have as many managers open as go workers as needed so that it can handle multiple service
requests simultaneously. The API service starts a manager in a GO worker routine and
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
	"golang.org/x/sync/errgroup"
)

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// loadEnvironment retrieves required settings from the environment
func (mngr *Manager) loadEnvironment() (mysqlCredentials, cbCredentials director.DbCredentials, err error) {
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

// Close is required after we are finished with a Manager. It usually recommended to
// call it with defer.
func (mngr *Manager) Close() error {
	return mngr.cb.Cluster.Close(nil)
}

// getCouchbaseConnection establishes the couchbase connection
// mysql connections are maintained in the mysql_director
func (mngr *Manager) getCouchbaseConnection(cbCredentials director.DbCredentials) (err error) {
	options := gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cbCredentials.User,
			Password: cbCredentials.Password,
		},
	}
	if err = options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment); err != nil {
		return fmt.Errorf("manager gocb ApplyProfile error: %w", err)
	}
	// Initialize the Connection
	var cluster *gocb.Cluster
	cluster, err = gocb.Connect("couchbase://"+cbCredentials.Host, options)
	if err != nil {
		return fmt.Errorf("manager gocb Connect error: %w", err)
	}
	mngr.cb.Cluster = cluster
	mngr.cb.Bucket = cluster.Bucket(cbCredentials.Bucket)
	err = mngr.cb.Bucket.WaitUntilReady(50*time.Second, nil)
	if err != nil {
		return fmt.Errorf("manager bucket.WaitUntilReady error: %w", err)
	}
	mngr.cb.Scope = mngr.cb.Bucket.Scope(cbCredentials.Scope)
	mngr.cb.Collection = mngr.cb.Bucket.Collection(cbCredentials.Collection)
	return nil
}

// upsertSubDocument updates a Couchbase subdocument
func (mngr *Manager) upsertSubDocument(path string, subDoc interface{}) error {
	mops := []gocb.MutateInSpec{
		gocb.UpsertSpec(path, subDoc, &gocb.UpsertSpecOptions{}),
	}
	upsertResult, err := mngr.cb.Collection.MutateIn(mngr.documentID, mops, &gocb.MutateInOptions{
		Timeout: 10050 * time.Millisecond,
	})
	if err != nil {
		return fmt.Errorf("manager upsertSubDocument error: %w", err)
	}
	// There is probably a better way to do this
	if upsertResult.MutationToken().BucketName() != "vxdata" {
		return fmt.Errorf("manager upsertSubDocument result bad upsertResult")
	}
	return nil
}

// getSubDocument retrieves a Couchbase subdocument
func (mngr *Manager) getSubDocument(path string, subDocPtr *interface{}) error {
	ops := []gocb.LookupInSpec{
		gocb.GetSpec(path, &gocb.GetSpecOptions{IsXattr: false}),
	}
	getResult, err := mngr.cb.Collection.LookupIn(mngr.documentID, ops, &gocb.LookupInOptions{})
	if err != nil {
		return fmt.Errorf("manager getSubDocument LookupIn error %w", err)
	}
	err = getResult.ContentAt(0, subDocPtr)
	if err != nil {
		return fmt.Errorf("manager getSubDocument getResult error %w", err)
	}
	return nil
}

// retrieve the results.blocks section of the document by subdoc get
func (mngr *Manager) getBlocks() (map[string]interface{}, error) {
	var blocks interface{}
	err := mngr.getSubDocument("results.blocks", &blocks)
	if err != nil {
		return nil, fmt.Errorf("manager getBlocks error %w", err)
	}
	return blocks.(map[string]interface{}), err
}

// retrieve the queryMap.blocks section of the document by subdoc get
func (mngr *Manager) getQueryBlocks() (map[string]interface{}, error) {
	var blocks interface{}
	err := mngr.getSubDocument("queryMap.blocks", &blocks)
	if err != nil {
		return nil, fmt.Errorf("manager getQueryBlocks error %w", err)
	}
	return blocks.(map[string]interface{}), err
}

// retrieve the PlotParams section of the document by subdoc get
func (mngr *Manager) getPlotParams() (map[string]interface{}, error) {
	var plotParams interface{}
	err := mngr.getSubDocument("plotParams", &plotParams)
	if err != nil {
		return nil, fmt.Errorf("manager getPlotParams error %w", err)
	}
	return plotParams.(map[string]interface{}), err
}

// retrieve the PlotParam.curves (this is an array) section of the document by subdoc get
func (mngr *Manager) getPlotParamCurves() ([]map[string]interface{}, error) {
	var curves interface{}
	var curveArray []map[string]interface{}
	err := mngr.getSubDocument("plotParams.curves", &curves)
	if err != nil {
		return nil, fmt.Errorf("manager getPlotParamCurves error %w", err)
	}
	for _, c := range curves.([]interface{}) {
		curveArray = append(curveArray, c.(map[string]interface{}))
	}
	return curveArray, err
}

// retrieve the dateRange section of the document by subdoc get
// and convert it to a dateRange struct
func (mngr *Manager) getDateRange() (director.DateRange, error) {
	var datesStr interface{}
	err := mngr.getSubDocument("dateRange", &datesStr)
	var dateRange director.DateRange
	// parse the daterange string
	// "02/19/2023 20:00 - 03/21/2023 20:00"
	if err != nil {
		return dateRange, fmt.Errorf("manager getDateRange error %w", err)
	}
	dateParts := strings.Split(datesStr.(string), " - ")
	fromTime, err := time.Parse("01/02/2006 15:04", dateParts[0])
	if err != nil {
		return dateRange, fmt.Errorf("manager getDataRange error converting from date to epoch error %w", err)
	}
	fromSecs := fromTime.Unix()
	if err != nil {
		return dateRange, fmt.Errorf("manager getDataRange error getting date range from document error %w", err)
	}
	toTime, err := time.Parse("01/02/2006 15:04", dateParts[1])
	if err != nil {
		return dateRange, fmt.Errorf("manager getDataRange error converting from date to epoch error %w", err)
	}
	toSecs := toTime.Unix()
	dateRange.FromSecs = fromSecs
	dateRange.ToSecs = toSecs
	return dateRange, err
}

// convertStdToPercent converts a standard deviation to a percent error
func (mngr *Manager) convertStdToPercent(std string) (percent float64, err error) {
	stdfloat, err := strconv.ParseFloat(std, 64)
	if err != nil {
		err = fmt.Errorf("manager convertStdToPercent error converting standard deviation %q to percent error: %w", std, err)
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

// getThresholds extracts the major and minor thresholds
func (mngr *Manager) getThresholds(plotParams map[string]interface{}) (minorThreshold, majorThreshold float64, err error) {
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
		minorThreshold, err = mngr.convertStdToPercent(plotParams["minor-threshold-by-stdv"].(string))
		if err != nil {
			return minorThreshold, majorThreshold, err
		}
		majorThreshold, err = mngr.convertStdToPercent(plotParams["major-threshold-by-stdv"].(string))
		if err != nil {
			return minorThreshold, majorThreshold, err
		}
	default:
		return minorThreshold, majorThreshold, fmt.Errorf("manager Run error getting threshold percentages")
	}
	return minorThreshold, majorThreshold, nil
}

// notifyMatsRefreash notifies the MATS scorecard app that a particular docID has been updated
func (mngr *Manager) notifyMatsRefresh(scorecardAppURL, docID string) error {
	err := client.NotifyScorecard(scorecardAppURL, docID)
	if err != nil {
		return fmt.Errorf("manager notifyMATSRefresh error: %w", err)
	}
	return err
}

func (mngr *Manager) processRegion(
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
	documentScorecardAppURL string,
	cellCountPtr *int,
) error {
	if strings.ToUpper(appName) == "CB" {
		return fmt.Errorf("Couchbase director is unimplemented")
	}
	// launch mysql director
	mysqlDirector, err := director.GetDirector("MysqlDirector", mysqlCredentials, dateRange, minorThreshold, majorThreshold)
	if err != nil {
		return fmt.Errorf("manager Run error getting director: %w", err)
	}
	defer mysqlDirector.CloseDB()

	*region, err = mysqlDirector.Run(queryRegionName, *region, queryRegion, cellCountPtr)
	if err != nil {
		return fmt.Errorf("manager Run error running director: %w", err)
	}

	err = mngr.upsertSubDocument(regionPath, region)
	if err != nil {
		return fmt.Errorf("manager Run error upserting resultRegion: %q error: %w", blockRegionName, err)
	}

	// notify server to update with scorecardApUrl
	// try to get the SCORECARD_APP_URL from the environment
	scorecardAppURL := os.Getenv("DEBUG_SCORECARD_APP_URL")
	if scorecardAppURL == "" {
		// not in environment. so use the one from the document
		scorecardAppURL = documentScorecardAppURL
	}
	err = mngr.notifyMatsRefresh(scorecardAppURL, mngr.documentID)
	if err != nil {
		return fmt.Errorf("manager Run error Failed to Notify appUrl %q: error: %w", scorecardAppURL, err)
	}
	return nil
}

// Run processes the docID associated with the manager
func (mngr *Manager) Run() (err error) {
	// load the environment
	cellCount := 0
	start := time.Now()
	// initially unknown
	mysqlCredentials, cbCredentials, err := mngr.loadEnvironment()
	if err != nil {
		return fmt.Errorf("manager loadEnvironmant error %w", err)
	}
	err = mngr.getCouchbaseConnection(cbCredentials)
	if err != nil {
		return fmt.Errorf("manager Run GetConnection error: %w", err)
	}
	defer mngr.cb.Cluster.Close(nil)
	// from here on we should be able to set an error status in the document, if we need to
	resultsBlocks, err := mngr.getBlocks()
	if err != nil {
		_ = mngr.SetStatus("error")
		return fmt.Errorf("manager Run error getting resultsBlocks: %w", err)
	}
	blockKeys := getMapKeys(resultsBlocks)
	sort.Strings(blockKeys)
	// get the appUrl from the first block - they should all be the same
	scorecardAppUrl := resultsBlocks[blockKeys[0]].(map[string]interface{})["blockApplication"].(string)
	// from this point on, we can notify the scorecard app with the status and error
	queryBlocks, err := mngr.getQueryBlocks()
	if err != nil {
		_ = mngr.SetStatus("error")
		err := fmt.Errorf("manager Run error getting queryBlocks: %w", err)
		_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
		return err
	}
	plotParams, err := mngr.getPlotParams()
	if err != nil {
		err := fmt.Errorf("manager Run error getting plotParamCurves: %w", err)
		_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
		_ = mngr.SetStatus("error")
		return err
	}
	minorThreshold, majorThreshold, err := mngr.getThresholds(plotParams)
	if err != nil {
		err := fmt.Errorf("manager Run error getting thresholds: %w", err)
		_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
		_ = mngr.SetStatus("error")
		return err
	}
	curves, err := mngr.getPlotParamCurves()
	if err != nil {
		err := fmt.Errorf("manager Run error getting plotParamCurves: %w", err)
		_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
		_ = mngr.SetStatus("error")
		return err
	}
	dateRange, err := mngr.getDateRange()
	if err != nil {
		err := fmt.Errorf("manager Run error getting daterange: %w", err)
		_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
		_ = mngr.SetStatus("error")
		return err
	}
	numCurves := len(curves)
	// blocks and queryBlocks have the same keys
	numBlocks := len(blockKeys)
	// create an errgroup for running all the block/regions in go routines
	errGroup := new(errgroup.Group)
	// don't really care what SINGLETHREADEDMANAGER env var is set to, just if it is set
	_, singleThreadedManager := os.LookupEnv("SINGLETHREADEDMANGER")
	if singleThreadedManager {
		log.Print("manager is Running SINGLETHREADEDMANGER")
	}
	for i := 0; i < numBlocks; i++ {
		blockName := blockKeys[i]
		block := resultsBlocks[blockName]
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
		blockRegionNames := getMapKeys(block.(map[string]interface{})["data"].(map[string]interface{}))
		sort.Strings(blockRegionNames)
		queryRegionNames := getMapKeys(queryData)
		sort.Strings(queryRegionNames)
		numBlockRegions := len(blockRegionNames)
		numQueryRegions := len(queryRegionNames)
		if numBlockRegions != numQueryRegions {
			err := fmt.Errorf("manager Run Number of block regions %v does not equal the number of query regions %v", numBlockRegions, numQueryRegions)
			_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
			_ = mngr.SetStatus("error")
			return err
		}
		if !reflect.DeepEqual(blockRegionNames, queryRegionNames) {
			err := fmt.Errorf("manager block regions list %v does not equal query regions list %v", blockRegionNames, queryRegionNames)
			_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
			_ = mngr.SetStatus("error")
			return err
		}
		for i := 0; i < numBlockRegions; i++ {
			queryRegionName := queryRegionNames[i]
			queryRegion := queryData[queryRegionName].(map[string]interface{})
			blockRegionName := blockRegionNames[i]
			var region interface{}
			regionPath := "results.blocks." + blockName + ".data." + blockRegionName
			err = mngr.getSubDocument(regionPath, &region)
			if err != nil {
				err := fmt.Errorf("error getting region SubDocument %w", err)
				_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
				_ = mngr.SetStatus("error")
				return err
			}
			if !singleThreadedManager {
				// process the region/block in the errgroup
				errGroup.Go(func() error {
					err = mngr.processRegion(
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
						scorecardAppUrl,
						&cellCount)
					return err
				})
			} else {
				err = mngr.processRegion(
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
					scorecardAppUrl,
					&cellCount)
				if err != nil {
					// set error in the status field
					err := fmt.Errorf("error processing scorecard single threaded Run %w", err)
					// set error status in document
					_ = mngr.SetStatus("error")
					_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
					return err
				}
			}
		}
	}
	if !singleThreadedManager {
		// Wait for all processRegions to complete, capture their error values
		err = errGroup.Wait()
		if err != nil {
			// set error in the status field
			err := fmt.Errorf("error processing scorecard multithreaded Run %w", err)
			// set error status in document
			_ = mngr.SetStatus("error")
			_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
			return err
		}
	}
	// set processedAt to now
	err = mngr.SetProcessedAt()
	if err != nil {
		err := fmt.Errorf("error setting processedAt %w", err)
		_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "error", err)
		err = mngr.SetStatus("error")
		return err
	}
	elapsed := time.Since(start)
	log.Printf("This run processed: %v cells in %v", cellCount, elapsed)
	_ = client.NotifyScorecardStatus(scorecardAppUrl, mngr.documentID, "ready", err)
	// set status to ready
	err = mngr.SetStatus("ready")
	return nil
}

// SetStatus updates the couchbase scorecard document with the processing status
func (mngr *Manager) SetStatus(status string) (err error) {
	stmnt := "UPDATE vxdata._default.SCORECARD SET status = \"" + status + "\" where meta().id=\"" + mngr.documentID + "\";"
	_, err = mngr.cb.Cluster.Query(stmnt, &gocb.QueryOptions{Adhoc: true})
	if err != nil {
		return err
	}
	return nil
}

// SetProcessedAt updates the couchbase scorecard document with the processed timestamp
func (mngr *Manager) SetProcessedAt() (err error) {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	stmnt := fmt.Sprintf("UPDATE vxdata._default.SCORECARD SET processedAt = %v where meta().id='%s';", timeStamp, mngr.documentID)
	_, err = mngr.cb.Cluster.Query(stmnt, &gocb.QueryOptions{Adhoc: true})
	if err != nil {
		return err
	}
	return nil
}

// newScorecardManager creates a correctly initialized scorecard manager. GetManager should be used by clients instead of this.
func newScorecardManager(documentID string) (*Manager, error) {
	scMgr := Manager{}
	scMgr.cb = &cbConnection{}
	scMgr.documentID = documentID
	return &scMgr, nil
}
