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
	"reflect"
	"strings"
	"time"
	_ "github.com/go-sql-driver/mysql"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
)

// the special sql types i.e. sql.NullInt64 sql.NullFloat64 are for handling null values
// see http://go-database-sql.org/nulls.html
// type CTCQueryRecord struct {
// 	avtime sql.NullInt32
// 	hit  sql.NullInt32
// 	miss sql.NullInt32
// 	fa   sql.NullInt32
// 	cn   sql.NullInt32
// }
// type CTCQueryRecords = []CTCQueryRecord

// type ScalarQueryRecord struct {
// 	avtime          sql.NullInt64
// 	squareDiffSum   sql.NullFloat64
// 	NSum            sql.NullInt32
// 	obsModelDiffSum sql.NullFloat64
// 	modelSum        sql.NullFloat64
// 	obsSum          sql.NullFloat64
// 	absSum          sql.NullFloat64
// }
// type ScalarQueryRecords = []ScalarQueryRecord
// type PreCalcQueryRecord struct {
// 	avtime  sql.NullInt64
// 	stat sql.NullFloat64
// }
// type PreCalcQueryRecords = []PreCalcQueryRecord

//type record *struct{}
//type records []record

var gp GoodnessPolarity
var minorThreshold Threshold
var majorThreshold Threshold

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

func NewMysqlDirector(mysqlCredentials DbCredentials) (*Director, error) {
	var db, err = getMySqlConnection(mysqlCredentials)
	if err != nil {
		return nil, fmt.Errorf("mysql_director NewMysqlDirector error: %q", err)
	} else {
		mysqlDirector.db = db
	}
	return &mysqlDirector, nil
}

func queryDataPreCalc(stmnt string, queryResult PreCalcRecords) (err error) {
	var rows *sql.Rows
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %q", stmnt)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var avtime int64
		var stat float64
		err = rows.Scan(&avtime, &stat)
		if err == nil {
			record := PreCalcRecord{avtime:avtime,stat:stat}
			queryResult = append(queryResult, record)
		} else {
			err = fmt.Errorf("mysqlDirector.Query error reading PreCalcRecord row %q", err)
			return err
		}
	}
	return nil
}

func queryDataCTC(stmnt string, queryResult CTCRecords) (err error) {
	var rows *sql.Rows
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %q", stmnt)
		return err
	}
	defer rows.Close()
	var record CTCRecord
	for rows.Next() {
		err = rows.Scan(record.avtime, record.hit, record.miss, record.fa, record.cn)
		if err != nil {
				queryResult = append(queryResult, record)
		} else {
			err = fmt.Errorf("mysqlDirector.Query error reading CTCRecord row %q", err)
			return err
		}
	}
	return nil
}

func queryDataScalar(stmnt string, queryResult ScalarRecords) (err error) {
	var rows *sql.Rows
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %q", stmnt)
		return err
	}
	defer rows.Close()
	var record ScalarRecord
	for rows.Next() {
		err = rows.Scan(record.avtime, record.squareDiffSum, record.NSum, record.obsModelDiffSum, record.modelSum, record.obsSum, record.absSum)
		if err != nil {
			queryResult = append(queryResult, record)
		} else {
			err = fmt.Errorf("mysqlDirector.Query error reading ScalarRecord row %q", err)
			return err
		}
	}
	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func processSub(resultElem ScorecardBlock, queryElem ScorecardBlock, statistics []string) error {
	var statisticType string
	var elemName = reflect.TypeOf(resultElem).Name()
	if contains(statistics, elemName) {
		statisticType = elemName
	}
	if reflect.TypeOf(resultElem).String() == "struct" {
		// elem is a cell
		// get the queries
		var ctlQueryStatement string = queryElem["controlQueryTemplate"].(string)
		var expQueryStatement string = queryElem["experimentalQueryTemplate"].(string)
		var ctlDataType string
		var expDataType string
		var err error
		var ctlQueryResult interface{}
		var expQueryResult interface{}
		// what kind of data?
		if strings.Contains(ctlQueryStatement, "hits") {
			ctlQueryResult = new(CTCRecords)
			expQueryResult = new(CTCRecords)
			// get the data
			err = queryDataCTC(ctlQueryStatement, ctlQueryResult.(CTCRecords))
			// handle error
			if err != nil {
				err = fmt.Errorf("mysql_director processSub error querying for ctlData - statement %q", ctlQueryStatement)
				return err
			}
			err = queryDataCTC(expQueryStatement, expQueryResult.(CTCRecords))
			if err != nil {
				err = fmt.Errorf("mysql_director processSub error querying for expData - statement %q", expQueryStatement)
				return err
			}
		} else if strings.Contains(ctlQueryStatement, "square_diff_sum") {
			ctlQueryResult = new(ScalarRecords)
			expQueryResult = new(ScalarRecords)
			// get the data
			err = queryDataScalar(ctlQueryStatement, ctlQueryResult.(ScalarRecords))
			// handle error
			if err != nil {
				err = fmt.Errorf("mysql_director processSub error querying for ctlData - statement %q", ctlQueryStatement)
				return err
			}
			err = queryDataScalar(expQueryStatement, expQueryResult.(ScalarRecords))
			if err != nil {
				err = fmt.Errorf("mysql_director processSub error querying for expData - statement %q", expQueryStatement)
				return err
			}
		} else if strings.Contains(ctlQueryStatement, "stat") {
			ctlQueryResult = new(PreCalcRecords)
			expQueryResult = new(PreCalcRecords)
			// get the data
			err = queryDataPreCalc(ctlQueryStatement, ctlQueryResult.(PreCalcRecords))
			// handle error
			if err != nil {
				err = fmt.Errorf("mysql_director processSub error querying for ctlData - statement %q", ctlQueryStatement)
				return err
			}
			err = queryDataPreCalc(expQueryStatement, expQueryResult.(PreCalcRecords))
			if err != nil {
				err = fmt.Errorf("mysql_director processSub error querying for expData - statement %q", expQueryStatement)
				return err
			}
		} else {
			return fmt.Errorf("mysql_director processSub unknown dataType for query %q", ctlQueryStatement)
		}
		// handle error
		if expDataType != ctlDataType {
			err = fmt.Errorf("mysql_director processSub ctlDataType %q does not equal expDataType %q", ctlQueryStatement, expQueryStatement)
			return err
		}
		// for all the input elements
		// build the input data elements - derive the statistic and summary value
		// for this element i.e. this cell in the scorecard
		// The build will fill in the value (write into the result)
		//Build(qr QueryResult, statisticType string, dataType string
		scc := NewTwoSampleTTestBuilder()
		var queryResult BuilderGenericResult = BuilderGenericResult{CtlData: ctlQueryResult, ExpData:expQueryResult}
		err = scc.Build(queryResult, statisticType)
		if err != nil {
			return fmt.Errorf("mysql_director error in ProcessSub %q", err)
		}
	} else {
		var keys []string = Keys(resultElem)
		for _, elemKey := range keys {
			var queryElem = queryElem[elemKey]
			var resultElem = resultElem[elemKey]
			return processSub(resultElem.(ScorecardBlock), queryElem.(ScorecardBlock), statistics)
		}
	}
	return nil
}

// build a section of a scorecard - this is a region (think vertical slice on the scorecard)
func (director *Director) Run(regionMap ScorecardBlock, queryMap ScorecardBlock) error {
	// This is recursive. Recurse down to the cell levl then traverse back up processing
	// all the cells on the way
	// get all the statistic strings (they are the keys of the regionMap)
	var statistics []string = make([]string, 0, len(regionMap))
	for k := range regionMap {
		statistics = append(statistics, k)
	}
	// process the regionMap (all the values will be filled in)
	err := processSub(regionMap, queryMap, statistics)
	if err != nil {
		return fmt.Errorf("mysql_director error in Run %q", err)
	}

	// manager will upsert the document
	return nil
}
