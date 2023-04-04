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
	"context"
	"database/sql"
	"fmt"
	"log"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

var dateRange DateRange

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
	var dataSource = fmt.Sprintf("%s:%s@tcp(%s)/", mysqlCredentials.User, mysqlCredentials.Password, mysqlCredentials.Host)
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

func NewMysqlDirector(mysqlCredentials DbCredentials, dateRange DateRange, minorThreshold float64, majorThreshold float64) (*Director, error) {
	var db, err = getMySqlConnection(mysqlCredentials)
	if err != nil {
		return nil, fmt.Errorf("mysql_director NewMysqlDirector error: %q", err)
	} else {
		mysqlDirector.db = db
		mysqlDirector.mysqlCredentials = mysqlCredentials
		mysqlDirector.queryBlock = ScorecardBlock{}
		mysqlDirector.resultBlock = ScorecardBlock{}
		mysqlDirector.dateRange = dateRange
		mysqlDirector.minorThreshold = minorThreshold
		mysqlDirector.majorThreshold = majorThreshold
	}
	return &mysqlDirector, nil
}

func queryDataPreCalc(stmnt string) (queryResult builder.PreCalcRecords, err error) {
	var rows *sql.Rows
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %q", err)
		return queryResult, err
	}
	defer rows.Close()
	var record builder.PreCalcRecord
	for rows.Next() {
		err = rows.Scan(&record.Avtime, &record.Stat)
		if err != nil {
			err = fmt.Errorf("mysqlDirector.Query error reading PreCalcRecord row %q", err)
			return queryResult, err
		} else {
			queryResult = append(queryResult, record)
		}
	}
	return queryResult, nil
}

func queryDataCTC(stmnt string) (queryResult builder.CTCRecords, err error) {
	var rows *sql.Rows
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %q", err)
		return queryResult, err
	}
	defer rows.Close()
	var record builder.CTCRecord
	for rows.Next() {
		err = rows.Scan(&record.Avtime, &record.Hit, &record.Miss, &record.Fa, &record.Cn)
		if err != nil {
			err = fmt.Errorf("mysqlDirector.Query error reading CTCRecord row %q", err)
			return queryResult, err
		} else {
			queryResult = append(queryResult, record)
		}
	}
	return queryResult, nil
}

// func queryDataScalar(stmnt string, queryResult builder.ScalarRecords) (err error) {
func queryDataScalar(stmnt string) (queryResult builder.ScalarRecords, err error) {
	var rows *sql.Rows
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %q", err)
		return queryResult, err
	}
	defer rows.Close()
	var record builder.ScalarRecord
	for rows.Next() {
		err = rows.Scan(&record.Avtime, &record.SquareDiffSum, &record.NSum, &record.ObsModelDiffSum, &record.ModelSum, &record.ObsSum, &record.AbsSum)
		if err != nil {
			err = fmt.Errorf("mysqlDirector.Query error reading ScalarRecord row %q", err)
			return queryResult, err
		} else {
			queryResult = append(queryResult, record)
		}
	}
	return queryResult, nil
}

// utility to test if an []string contains a specific string
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

var statistics []string
var statisticType string
var thisIsALeaf bool

// Recursively process a region/Block until all the leaves (which are cells) have been traversed and processed
func processSub(region interface{}, queryElem interface{}) (interface{}, error) {
	thisIsALeaf = false
	var err error
	keys := Keys(queryElem.(map[string]interface{}))
	if contains(keys, "controlQueryTemplate") {
		thisIsALeaf = true
	}
	if thisIsALeaf { // now we have a struct
		// if I already had a leaf on this branch trim it
		// get the queries
		var ctlQueryStatement string = queryElem.(map[string]interface{})["controlQueryTemplate"].(string)
		var expQueryStatement string = queryElem.(map[string]interface{})["experimentalQueryTemplate"].(string)
		// substitute the {{fromSecs}} and {{toSecs}}
		ctlQueryStatement = strings.Replace(ctlQueryStatement, "{{fromSecs}}", fmt.Sprint(dateRange.FromSecs), -1)
		ctlQueryStatement = strings.Replace(ctlQueryStatement, "{{toSecs}}", fmt.Sprint(dateRange.ToSecs), -1)
		expQueryStatement = strings.Replace(expQueryStatement, "{{fromSecs}}", fmt.Sprint(dateRange.FromSecs), -1)
		expQueryStatement = strings.Replace(expQueryStatement, "{{toSecs}}", fmt.Sprint(dateRange.ToSecs), -1)
		var err error
		var queryResult interface{}
		queryError := false

		// what kind of data?
		if strings.Contains(ctlQueryStatement, "hits") {
			// get the data
			ctlQueryResult, err := queryDataCTC(ctlQueryStatement)
			// handle error
			if err != nil {
				queryError = true
				log.Printf("mysql_director queryDataCTC ctlQueryStatement error %q", err)
			} else {
				expQueryResult, err := queryDataCTC(expQueryStatement)
				if err != nil {
					queryError = true
					log.Printf("mysql_director queryDataCTC expQueryStatement error %q", err)
				} else {
					queryResult = builder.BuilderCTCResult{CtlData: ctlQueryResult, ExpData: expQueryResult}
				}
			}
		} else if strings.Contains(ctlQueryStatement, "square_diff_sum") {
			// get the data
			ctlQueryResult, err := queryDataScalar(ctlQueryStatement)
			// handle error
			if err != nil {
				queryError = true
				log.Printf("mysql_director queryDataScalar ctlQueryStatement error %q", err)
			} else {
				expQueryResult, err := queryDataScalar(expQueryStatement)
				if err != nil {
					queryError = true
					log.Printf("mysql_director queryDataScalar expQueryStatementerror %q", err)
				} else {
					queryResult = builder.BuilderScalarResult{CtlData: ctlQueryResult, ExpData: expQueryResult}
				}
			}
		} else if strings.Contains(ctlQueryStatement, "stat") {
			// get the data
			ctlQueryResult, err := queryDataPreCalc(ctlQueryStatement)
			// handle error
			if err != nil {
				queryError = true
				log.Printf("mysql_director queryDataPreCalc ctlQueryStatement error %q", err)
			} else {
				expQueryResult, err := queryDataPreCalc(expQueryStatement)
				if err != nil {
					queryError = true
					log.Printf("mysql_director queryDataPreCalc expQueryStatement error %q", err)
				} else {
					queryResult = builder.BuilderPreCalcResult{CtlData: ctlQueryResult, ExpData: expQueryResult}
				}
			}
		} else {
			// unknown data type
			log.Printf("mysql_director queryDataPreCalc error %v", err)
			return -9999, fmt.Errorf("mysql_director queryDataPreCalc error %q", err)
		}

		// for all the input elements
		// build the input data elements - derive the statistic and summary value
		// for this element i.e. this cell in the scorecard
		// The build will fill in the value (write into the result)
		//Build(qr QueryResult, statisticType string, dataType string
		if queryError {
			log.Printf("mysql_director query error %v", err)
			return -9999, err
		} else {
			scc := builder.NewTwoSampleTTestBuilder()
			value, err := (scc.Build(queryResult, statisticType, mysqlDirector.minorThreshold, mysqlDirector.majorThreshold))
			if err != nil {
				return -9999, fmt.Errorf("mysql_director processSub error from builder %q", err)
			} else {
				return int(value), nil
			}
		}
	} else {
		// this is a branch (not a leaf) so we keep traversing
		// check to see if this is a statistic elem, so we can set the statisticType
		var keys []string = Keys((region).(map[string]interface{}))
		for _, elemKey := range keys {
			if contains(statistics, elemKey) {
				statisticType = elemKey
			}
			var queryElem = queryElem.(map[string]interface{})[elemKey]
			region.(map[string]interface{})[elemKey], err = processSub(region.(map[string]interface{})[elemKey], queryElem)
			if err != nil {
				return -9999, err
			}
		}
	}
	return region, nil
}

// build a section of a scorecard - this is a region of a block (think vertical slice on the scorecard)
func (director *Director) Run(region interface{}, queryMap map[string]interface{}) (interface{}, error) {
	// This is recursive. Recurse down to the cell levl then traverse back up processing
	// all the cells on the way
	// get all the statistic strings (they are the keys of the regionMap)
	statistics = Keys((region).(map[string]interface{})) // declared at the top
	dateRange = director.dateRange
	// process the regionMap (all the values will be filled in)
	region, err := processSub(region, queryMap)
	if err != nil {
		return region, fmt.Errorf("mysql_director error in Run %q", err)
	}
	// manager will upsert the document
	return region, nil
}
