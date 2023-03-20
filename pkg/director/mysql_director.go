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
	"strings"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// the special sql types i.e. sql.NullInt64 sql.NullFloat64 are for handling null values
// see http://go-database-sql.org/nulls.html
type CTCRecord struct {
	hit sql.NullInt32
	miss sql.NullInt32
	fa sql.NullInt32
	cn sql.NullInt32
	time sql.NullInt32
}
type CTCRecords = []*CTCRecord

type ScalarRecord struct {
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
	value sql.NullFloat64
	time sql.NullInt64
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

func queryData(stmnt string, queryResultDataPtr *[]interface{}) (dataType string, err error) {
    var rows* sql.Rows
	// There is undoubtedly a better way to do this!
	dataType = "PreCalcRecord"
	if strings.Contains(stmnt, "hit") {
		dataType = "CTCRecord"
	}
	if strings.Contains(stmnt,"squareDiffSum") {
		dataType = "ScalarRecord"
	}
	rows, err = mysqlDirector.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %q", stmnt)
		return dataType, err
	  }
	defer rows.Close()
	switch dataType {
	case "PreCalcRecord":
		var recordPtr *PreCalcRecord = new(PreCalcRecord)
		for rows.Next() {
			err = rows.Scan(recordPtr.value, recordPtr.time)
			if err != nil {
				if recordPtr.value.Valid && recordPtr.time.Valid {
					*queryResultDataPtr = append(*queryResultDataPtr, recordPtr)
				}
			} else {
				err = fmt.Errorf ("mysqlDirector.Query error reading PreCalcRecord row %q", err)
				return dataType, err
			}
		}
		//data = records.(PreCalcRecords)

	case "CTCRecord":
		var recordPtr *CTCRecord = new(CTCRecord)
		for rows.Next() {
			err = rows.Scan(recordPtr.hit, recordPtr.miss, recordPtr.fa, recordPtr.cn, recordPtr.time)
			if err != nil {
				if recordPtr.hit.Valid && recordPtr.miss.Valid && recordPtr.fa.Valid && recordPtr.cn.Valid && recordPtr.time.Valid {
					*queryResultDataPtr = append(*queryResultDataPtr, recordPtr)
				}
			} else {
				err = fmt.Errorf ("mysqlDirector.Query error reading CTCRecord row %q", err)
				return dataType, err
			}
		}
	case "ScalarRecord":
		var recordPtr *ScalarRecord = new(ScalarRecord)
		for rows.Next() {
			err = rows.Scan(recordPtr.squareDiffSum, recordPtr.NSum, recordPtr.obsModelDiffSum, recordPtr.modelSum, recordPtr.obsSum, recordPtr.absSum, recordPtr.time)
			if err != nil {
				// check for NULL values
				if recordPtr.squareDiffSum.Valid &&  recordPtr.NSum.Valid && recordPtr.obsModelDiffSum.Valid && recordPtr.modelSum.Valid && recordPtr.obsSum.Valid && recordPtr.absSum.Valid && recordPtr.time.Valid {
					*queryResultDataPtr = append(*queryResultDataPtr, recordPtr)
				}
			} else {
				err = fmt.Errorf ("mysqlDirector.Query error reading ScalarRecord row %q", err)
				return dataType, err
			}
		}
	}
	// should be nil
	return dataType, err
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func processSub(resultElem ScorecardBlock, queryElem ScorecardBlock, statistics []string) error{
	var statisticType string
	var elemName = reflect.TypeOf(resultElem).Name()
	if contains(statistics,elemName) {
		statisticType = elemName
	}
	if reflect.TypeOf(resultElem).String() == "struct" {
		// elem is a cell
		// get the queries
		var ctlQueryStatement string = queryElem["controlQueryTemplate"].(string)
		var expQueryStatement string = queryElem["experimentalQueryTemplate"].(string)
		var queryResultPtr *builder.QueryResult = new(builder.QueryResult)
		var ctlDataType string
		var expDataType string
		var dataType string
		var err error
		// get the data
		ctlDataType, err = queryData(ctlQueryStatement, &(*queryResultPtr.CtlData))
		// handle error
		if err != nil {
			err = fmt.Errorf("mysql_director processSub error querying for ctlData - statement %q", ctlQueryStatement)
			return err
		}
		expDataType, err = queryData(expQueryStatement,&(*queryResultPtr.ExpData))
		if err != nil {
			err = fmt.Errorf("mysql_director processSub error querying for expData - statement %q", expQueryStatement)
			return err
		}
		// handle error
		if expDataType != ctlDataType {
			err = fmt.Errorf("mysql_director processSub ctlDataType %q does not equal expDataType %q", ctlQueryStatement, expQueryStatement)
			return err
		}
		dataType = expDataType
		// for all the input elements
		// build the input data elements - derive the statistic and summary value
		// for this element i.e. this cell in the scorecard
		// The build will fill in the value (write into the result)
		//Build(qr QueryResult, statisticType string, dataType string
		scc := builder.NewTwoSampleTTestBuilder()
		scc.Build(queryResultPtr, statisticType, dataType)
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
func Run(regionMap ScorecardBlock, queryMap ScorecardBlock) error {
	// This is recursive. Recurse down to the cell level then traverse back up processing
	// all the cells on the way
	// get all the statistic strings (they are the keys of the regionMap)
	var statistics []string = make([]string, 0, len(regionMap))
	for k := range regionMap {
		statistics = append(statistics, k)
	}
    // process the regionMap (all the values will be filled in)
	processSub(regionMap, queryMap, statistics)
	// upsert the document
	return nil
}
