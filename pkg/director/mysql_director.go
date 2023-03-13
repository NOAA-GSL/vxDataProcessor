package director

/*
This Director has the following responsibilities...
1. Receive an app URL and a pointer to an sc_row (which is a map).
2. Query the app for the mysql query template.
3. Create a query from the template by substituting the necessary variables into the template
(these are embedded in the scorecard row).
4. Retrieve the input data.
5. Format the input data into the proper structures for the builders.
An InputData structure has an array of values and an array of corresponding times for the experimental
data and also for the control data for a specific cell, a statistic and a pointer to the result
structure where the cell result value is to be placed.
7. Fire off builders in go worker routines to process all the cell DerivedDataElement structures
   1. the builder has to do these steps...
      1. Perform time matching on the input data
      2. Perform a statistic calculation (RMSE, BIAS, etc on the input data) and put it into a DerivedDataElement
	  using one of the statistic routines from builder_stats package.
      3. Compute the significance for the DerivedDataElement
      4. write the result value into the result structure. (value is a pointer)
	  5. politely die and go away.
*/


import (
	"fmt"
	"time"
	"reflect"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder_stats"
	"context"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)


// these should come from the scorecard

var gp builder.GoodnessPolarity
var minorThreshold builder.Threshold
var majorThreshold builder.Threshold

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getMySqlConnection(mysqlCredentials DbCredentials) (*sql.DB, error) {
	// get the connection
	var driver = "mysql"
	//user:password@tcp(localhost:5555)
	var dataSource  = fmt.Sprintf("%s:%s@tcp(%s)", mysqlCredentials.user, mysqlCredentials.password, mysqlCredentials.host)
	var db *sql.DB
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		return nil, fmt.Errorf("mysql_director getMySqlConnection sql open error %q", err)
	}
	var ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql_director Build sql open/ping error: %q", err) 
	}
	return db, nil
}

var mysqlDirector = Director{}
func NewMysqlDirector(mysqlCredentials DbCredentials) (*Director, error) {
	var db, err = getMySqlConnection(mysqlCredentials)
	if err != nil {
		return nil, fmt.Errorf("mysql_director NewMysqlDirector error: %q", err)
	} else {
		mysqlDirector.db = db
	}
	return &mysqlDirector, nil
}

func queryData(stmnt string, data *struct{}) error {
	var err error
    var rows* sql.Rows
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		// handle this error better than this
		panic(err)
	  }
	defer rows.Close()
	var columNames []string
	columNames, err = rows.Columns
	for _, v := range columNames {
		if v == "hits" {
			return true
		}
	}
	fmt.Println(columNames)
	for rows.Next() {
		err = rows.Scan(data)
		// handle error
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func processSub(resultElem ScorecardBlock, queryElem ScorecardBlock) error{
	if reflect.TypeOf(resultElem).String() == "struct" {
		// elem is a cell
		// get the queries
		var ctlQueryStatement = queryElem["controlQueryTemplate"]
		var expQueryStatement = queryElem["experimentalQueryTemplate"]
		var ctlData struct{}
		// get the data
		var err = queryData(ctlQueryStatement, &ctlData)
		// handle error
		if err != nil {
			panic(err)
		}
		var expData struct{}
		err = queryData(expQueryStatement, &ctlData)
		// handle error
		if err != nil {
			panic(err)
		}
		// for all the input elements
		// build the input data elements and
		return nil
	}
	var keys []string = Keys(resultElem)
	for _, elemKey := range keys {
		var queryElem = queryElem[elemKey]
		var resultElem = resultElem[elemKey]
		return processSub(resultElem.(ScorecardBlock), queryElem.(ScorecardBlock))
	}
	return nil
}

// build a section of a scorecard - this is a region (think vertical slice on the scorecard)
func Run(regionMap ScorecardBlock, queryMap ScorecardBlock) error {
	processSub(regionMap, queryMap)
	// upsert the document
	return nil
}
