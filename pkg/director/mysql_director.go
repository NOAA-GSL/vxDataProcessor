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
	"github.com/go-sql-driver/mysql"
	"database/sql"
)


// the special sql types i.e. sql.NullInt46 sql.NullFloat64 are for handling null values
// see http://go-database-sql.org/nulls.html
type CTCRecord = struct {
	hit sql.NullInt
	miss sql.NullInt
	fa sql.NullInt
	cn sql.NullInt
	time sql.NullInt64
}
type CTCRecords = []*CTCrecord

type ScalarRecord = struct {
	squareDiffSum sql.NullFloat64
	NSum sql.NullInt32
	obsModelDiffSum sql.NullFloat64
	modelSum sql.NullFloat64
	obsSum sql.NullFloat64
	absSum sql.NullFloat64
	time sql.NullInt64
}
type ScalarRecords = []*ScalarRecord
type PreCalcRecord struct {
	value NullFloat64
	time NullInt64
}
type PreCalcRecords = []*PreCalcRecord

type record *struct{}
type records []record

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
	var dataType string = "PreCalcRecord"
	var record record
	var records []records

	if strings.contains(stmnt, "hit") {
		dataType = "CTCRecord"
	}
	if strings.contains(stmnt,"squareDiffSum") {
		dataType = "ScalarRecord"
	}
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		// handle this error better than this
		panic(err)
	  }
	defer rows.Close()
	switch dataType {
	case "PreCalcRecord":
		record = &new(PreCalcRecord)
		records = make(PreCalcRecords)
		for rows.Next() {
			err = rows.Scan(&record.value, &record.timetime)
			if err != nil {
				if record.value.valid && record.time.valid {
					records.append(&record)
				}
			} else {
				log.Errorf ("mysqlDirector.Query error reading PreCalcRecord row %q", err)
			}
		}
	case "CTCRecord":
		record = new(CTCRecord)
		records = make(CTCRecords)
		for rows.Next() {
			err = rows.Scan(&record.hit, &record.miss, &record.fa, &record.cn, &record.time)
			if err != nil {
				if record.hit.valid && record.miss.valid && record.fa.valid && record.cn.valid && record.time.valid {
					records.append(&record)
				}
			} else {
				log.Errorf ("mysqlDirector.Query error reading CTCRecord row %q", err)
			}
		}
	case "ScalarRecord":
		record = new(ScalarRecord)
		records = make(ScalarRecords)
		for rows.Next() {
			err = rows.Scan(&record.squareDiffSum, &record.NSum, &record.obsModelDiffSum, &record.modelSum, &record.obsSum, &record.absSum, &record.time)
			if err != nil {
				// check for NULL values
				if record.squareDiffSum.valid &&  record.NSum.valid && record.obsModelDiffSum.valid && record.modelSum.valid && record.obsSum.valid && record.absSum.valid && record.time.valid {
					records.append(&record)
				}
			} else {
				log.Errorf ("mysqlDirector.Query error reading ScalarRecord row %q", err)
			}
		}
	}
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
