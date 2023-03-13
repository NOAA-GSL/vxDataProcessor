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
	"bufio"
	"errors"
	"fmt"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/director"
	"github.com/couchbase/gocb/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
	"time"
)



var mysqlCredentials director.DbCredentials

func loadEnvironmant(environmentFile string) error {
	err := godotenv.Load(environmentFile)
	if err != nil {
		return fmt.Errorf("Error loading .env file: %q", environmentFile)
	}
	var cbCredentials = director.DbCredentials{
		scope: "_default",
		collection: "SCORECARD",
		host: os.Getenv("CB_HOST"),
	}

	if cbCredentials.host == "" {
		return fmt.Errorf("Undefined CB_HOST in environment")
	}
	cbCredentials.user = os.Getenv("CB_USER")
	if cbCredentials.user == "" {
		return fmt.Errorf("Undefined CB_USER in environment")
	}
	cbCredentials.password = os.Getenv("CB_PASSWORD")
	if cbCredentials.password == "" {
		return fmt.Errorf("Undefined CB_PASSWORD in environment")
	}
	cbCredentials.bucket = os.Getenv("CB_BUCKET")
	if cbCredentials.bucket == "" {
		return fmt.Errorf("Undefined CB_BUCKET in environment")
	}

	// refer to https://github.com/go-sql-driver/mysql/#dsn-data-source-name
	mysqlCredentials.host = os.Getenv("MYSQL_HOST")
	if mysqlCredentials.host == "" {
		return fmt.Errorf("Undefined MYSQL_HOST in environment")
	}
	mysqlCredentials.user = os.Getenv("MYSQL_USER")
	if mysqlCredentials.user == "" {
		return fmt.Errorf("Undefined MYSQL_USER in environment")
	}
	mysqlCredentials.password = os.Getenv("MYSQL_PASSWORD")
	if mysqlCredentials.password == "" {
		return fmt.Errorf("Undefined MYSQL_PASSWORD in environment")
	}
}

// get the couchbase connection
// mysql connections are maintained in the mysql_director
func getConnection() (*cbConnection, error) {
	var cbCredentials, err = loadEnvironmant()
	if err != nil {
		return nil, fmt.Errorf("manager loadEnvironmant error %q", err)
	}

	var cbConnection cbConnection
	var options = gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cbCredentials.user,
			Password: cbCredentials.password,
		},
	}
	if err = options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment); err != nil {
		return nil, fmt.Errorf("manager gocb ApplyProfile error: %q", err)
	}
	// Initialize the Connection
	var cluster *gocb.Cluster
	cluster, err = gocb.Connect("couchbase://"+cbCredentials.host, options)
	if err != nil {
		return nil, fmt.Errorf("manager gocb Connect error: %q", err)
	}
	cbConnection.cluster = cluster
	cbConnection.bucket = cbConnection.cluster.Bucket(cbCredentials.bucket)
	err = cbConnection.bucket.WaitUntilReady(50*time.Second, nil)
	if err != nil {
		return nil, fmt.Errorf("manager bucket.WaitUntilReady error: %q", err)
	}
	cbConnection.scope = cbConnection.bucket.Scope(cbCredentials.scope)
	cbConnection.collection = cbConnection.bucket.Collection(cbCredentials.collection)
	return &cbConnection, nil
}


func (mngr *Manager) Run(documentId string, environmentFile string) error {
	// load the environment
	mngr.documentId = documentId
	mngr.environmentFile = environmentFile
	loadEnvironmant(mngr.environmentFile)
	// get the connection
	var cb *cbConnection
	cb , err = getConnection()
	if err != nil {
		return fmt.Errorf("manager Build GetConnection error: %q", err)
	}

	// get the scorecard document
	var scorecardDataIn *gocb.GetResult
	scorecardDataIn, err = cb.CB_collection.Get("documentId", nil)
	if err != nil {
		fmt.Errorf("manager Build error getting scorecard: %q  error: %q", documentId, err)
	}
	// get the unmarshalled document (the Content) from the result
	var scorecardCB interface{}
	err = scorecardDataIn.Content(&scorecardCB)
	if err != nil {
		fmt.Errorf("manager Build error getting scorecard Content: %q", err)
	}
	var appName string
	// iterate the rows in the scorecard
	var blocks = scorecardCB["SCORECARD"]["results"]["blocks"]
	var queryBlocks = scorecardCB["SCORECARD"]["queryMap"]["blocks"]
	for blockLabel := range Keys(blocks) { // what kind of app?
		for _,c := range scorecardCB["SCORECARD"]["plotParams"]["curves"] {
			//fmt.Println(i, s)
			if c.label == blockLabel {
				appName = c.application
			}
		}
		var regionLabels = director.Keys(blocks["data"])
		// launch a director for each region
		for _,regionLabel := range regionLabels {
			// launch a director for this region
			if strings.ToUpper(appName) == "CB" {
				// launch CB director - which we don't have yet
			} else {
				//launch mysql director
				var director = director.GetDirector("MysqlDirector", mysqlCredentials)
				director.run(blocks[BlockLabel]["data"][regionLabel], queryBlocks[BlockLabel]["data"][regionLabel])
			}
		}
		// upsert the document
		return nil
	}
}
