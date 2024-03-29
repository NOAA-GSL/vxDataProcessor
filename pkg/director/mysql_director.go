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
	"os"
	"strings"
	"sync"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	_ "github.com/go-sql-driver/mysql"
)

const (
	noTableFound   = "Error 1146 (42S02)"
	convertingNull = "converting NULL"
)

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getMySQLConnection establishes a connection to the given SQL database
// connection strings should be like: user:password@tcp(localhost:5555)
func (director *Director) getMySqlConnection(mysqlCredentials DbCredentials) (*sql.DB, error) {
	// get the connection
	driver := "mysql"
	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/", mysqlCredentials.User, mysqlCredentials.Password, mysqlCredentials.Host)
	var db *sql.DB
	db, err := sql.Open(driver, dataSource)
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	if err != nil {
		return nil, fmt.Errorf("mysql_director getMySqlConnection sql open error %w", err)
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql_director Build sql open/ping error: %w", err)
	}
	return db, nil
}

// newMySQLDirector creates a correctly initialized MySQL director. GetDirector should be used by clients instead of this.
func newMySQLDirector(mysqlCredentials DbCredentials, dateRange DateRange, minorThreshold, majorThreshold float64) (*Director, error) {
	mysqlDirector := Director{}
	db, err := mysqlDirector.getMySqlConnection(mysqlCredentials)
	if err != nil {
		return nil, fmt.Errorf("mysql_director NewMysqlDirector error: %w", err)
	} else {
		mysqlDirector.db = db
		mysqlDirector.mysqlCredentials = mysqlCredentials
		mysqlDirector.queryBlock = ScorecardBlock{}
		mysqlDirector.resultBlock = ScorecardBlock{}
		mysqlDirector.dateRange = dateRange
		mysqlDirector.minorThreshold = minorThreshold
		mysqlDirector.majorThreshold = majorThreshold
		mysqlDirector.wg = &sync.WaitGroup{}
	}
	return &mysqlDirector, nil
}

func (director *Director) queryDataPreCalc(stmnt string) (queryResult builder.PreCalcRecords, err error) {
	var rows *sql.Rows
	rows, err = director.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %w", err)
		return queryResult, err
	}
	defer rows.Close()
	var record builder.PreCalcRecord
	for rows.Next() {
		err = rows.Scan(&record.Avtime, &record.Stat)
		if err != nil {
			err = fmt.Errorf("mysqlDirector.Query error reading PreCalcRecord row %w", err)
			return queryResult, err
		} else {
			queryResult = append(queryResult, record)
		}
	}
	return queryResult, nil
}

func (director *Director) queryDataCTC(stmnt string) (queryResult builder.CTCRecords, err error) {
	var rows *sql.Rows
	rows, err = director.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %w", err)
		return queryResult, err
	}
	defer rows.Close()
	var record builder.CTCRecord
	for rows.Next() {
		err = rows.Scan(&record.Avtime, &record.Hit, &record.Miss, &record.Fa, &record.Cn)
		if err != nil {
			err = fmt.Errorf("mysqlDirector.Query error reading CTCRecord row %w", err)
			return queryResult, err
		} else {
			queryResult = append(queryResult, record)
		}
	}
	return queryResult, nil
}

// func queryDataScalar(stmnt string, queryResult builder.ScalarRecords) (err error) {
func (director *Director) queryDataScalar(stmnt string) (queryResult builder.ScalarRecords, err error) {
	var rows *sql.Rows
	rows, err = director.db.Query(stmnt)
	if err != nil {
		err = fmt.Errorf("mysql_director queryData Query failed: %w", err)
		return queryResult, err
	}
	defer rows.Close()
	var record builder.ScalarRecord
	for rows.Next() {
		err = rows.Scan(&record.Avtime, &record.SquareDiffSum, &record.NSum, &record.ObsModelDiffSum, &record.ModelSum, &record.ObsSum, &record.AbsSum)
		if err != nil {
			err = fmt.Errorf("mysqlDirector.Query error reading ScalarRecord row %w", err)
			return queryResult, err
		} else {
			queryResult = append(queryResult, record)
		}
	}
	return queryResult, nil
}

// used to return value and err from go routines
type errval struct {
	err error
	val int
}

var singleThreadedDirector bool = false

// Recursively process a region/Block until all the leaves (which are cells) have been traversed and processed
func (director *Director) processSub(queryRegionName string, region interface{}, queryElem interface{}, cellCountPtr *int, keychain *[]string, dateRange DateRange) (interface{}, error) {
	var err error
	keys := getMapKeys(queryElem.(map[string]interface{}))
	thisIsALeaf := false
	for _, k := range keys {
		if k == "controlQueryTemplate" {
			thisIsALeaf = true
			break
		}
	}
	if thisIsALeaf { // now we have a struct
		// log statement uncomment for debugging
		// log.Printf("mysql_director processSub leaf keys are %q", keys)

		// get the queries
		var ctlQueryStatement string = queryElem.(map[string]interface{})["controlQueryTemplate"].(string)
		var expQueryStatement string = queryElem.(map[string]interface{})["experimentalQueryTemplate"].(string)
		// substitute the {{fromSecs}} and {{toSecs}}
		ctlQueryStatement = strings.Replace(ctlQueryStatement, "{{fromSecs}}", fmt.Sprint(director.dateRange.FromSecs), -1)
		ctlQueryStatement = strings.Replace(ctlQueryStatement, "{{toSecs}}", fmt.Sprint(director.dateRange.ToSecs), -1)
		expQueryStatement = strings.Replace(expQueryStatement, "{{fromSecs}}", fmt.Sprint(director.dateRange.FromSecs), -1)
		expQueryStatement = strings.Replace(expQueryStatement, "{{toSecs}}", fmt.Sprint(director.dateRange.ToSecs), -1)
		var queryResult interface{}
		queryError := false

		// what kind of data?
		if strings.Contains(ctlQueryStatement, "hit") {
			// get the data
			ctlQueryResult, err := director.queryDataCTC(ctlQueryStatement)
			if len(ctlQueryResult) == 0 && err == nil {
				// no data is ok, but no need to go on either
				return builder.ErrorValue, nil
			}
			if err != nil {
				queryError = true
				if !strings.Contains(err.Error(), noTableFound) && !strings.Contains(err.Error(), convertingNull) {
					log.Printf("mysql_director queryDataCTC ctlQueryStatement error %q", err)
				}
			} else {
				expQueryResult, err := director.queryDataCTC(expQueryStatement)
				if len(expQueryResult) == 0 && err == nil {
					// no data is ok, but no need to go on either
					return builder.ErrorValue, nil
				}
				// handle error
				if err != nil {
					queryError = true
					if !strings.Contains(err.Error(), noTableFound) && !strings.Contains(err.Error(), convertingNull) {
						log.Printf("mysql_director queryDataCTC expQueryStatement error %q", err)
					}
				} else {
					queryResult = builder.BuilderCTCResult{CtlData: ctlQueryResult, ExpData: expQueryResult}
				}
			}
		} else if strings.Contains(ctlQueryStatement, "square_diff_sum") {
			// get the data
			ctlQueryResult, err := director.queryDataScalar(ctlQueryStatement)
			if len(ctlQueryResult) == 0 && err == nil {
				// no data is ok, but no need to go on either
				return builder.ErrorValue, nil
			}
			// handle error
			if err != nil {
				queryError = true
				if !strings.Contains(err.Error(), noTableFound) && !strings.Contains(err.Error(), convertingNull) {
					log.Printf("mysql_director queryDataScalar ctlQueryStatement error %q", err)
				}
			} else {
				expQueryResult, err := director.queryDataScalar(expQueryStatement)
				if len(expQueryResult) == 0 && err == nil {
					// no data is ok, but no need to go on either
					return builder.ErrorValue, nil
				}
				// handle error
				if err != nil {
					queryError = true
					if !strings.Contains(err.Error(), noTableFound) && !strings.Contains(err.Error(), convertingNull) {
						log.Printf("mysql_director queryDataScalar expQueryStatement error %q", err)
					}
				} else {
					queryResult = builder.BuilderScalarResult{CtlData: ctlQueryResult, ExpData: expQueryResult}
				}
			}
		} else if strings.Contains(ctlQueryStatement, "stat") {
			// get the data
			ctlQueryResult, err := director.queryDataPreCalc(ctlQueryStatement)
			if len(ctlQueryResult) == 0 && err == nil {
				// no data is ok, but no need to go on either
				return builder.ErrorValue, nil
			}
			// handle error
			if err != nil {
				queryError = true
				if !strings.Contains(err.Error(), noTableFound) && !strings.Contains(err.Error(), convertingNull) {
					log.Printf("mysql_director queryDataPreCalc ctlQueryStatement error %q", err)
				}
			} else {
				expQueryResult, err := director.queryDataPreCalc(expQueryStatement)
				if len(expQueryResult) == 0 && err == nil {
					// no data is ok, but no need to go on either
					return builder.ErrorValue, nil
				}
				if err != nil {
					queryError = true
					if !strings.Contains(err.Error(), noTableFound) && !strings.Contains(err.Error(), convertingNull) {
						log.Printf("mysql_director queryDataPreCalc expQueryStatement error %q", err)
					}
				} else {
					queryResult = builder.BuilderPreCalcResult{CtlData: ctlQueryResult, ExpData: expQueryResult}
				}
			}
		} else {
			// unknown data type
			return builder.ErrorValue, fmt.Errorf("mysql_director processSub error unknown data type - ctlQueryStatement %s - %w", ctlQueryStatement, err)
		}

		// for all the input elements
		// build the input data elements - derive the statistic and summary value
		// for this element i.e. this cell in the scorecard.
		// The actual scorecard data location for the cell is written in
		// the branch part of processSub when it encounters a leaf. See below.
		// Build(qr QueryResult, statisticType string, dataType string
		if queryError {
			if err != nil {
				log.Printf("mysql_director query error %v", err)
			}
			return builder.ErrorValue, err
		} else {
			valueStruct := builder.ValueStruct{}
			if !singleThreadedDirector {
				director.wg.Add(1)
				// run builder in parallel
				c := make(chan errval)
				go func(regionName string) {
					defer director.wg.Done()
					*cellCountPtr++
					scc := builder.NewTwoSampleTTestBuilder()
					_ = scc.SetKeyChain(*keychain) // ignore errorscc.
					value, err := (scc.Build(queryResult, director.statisticType, director.minorThreshold, director.majorThreshold))
					// remove this leaf key from the keychain
					if len(*keychain) > 0 {
						kc := *keychain
						kc = kc[:len(kc)-1]
						*keychain = kc
					}
					c <- errval{err: err, val: value}
					// build the value structure for this cell
					valueStruct.Path = scc.GetPath()
					valueStruct.GoodnessPolarity = scc.GetGoodnessPolarity()
					valueStruct.MajorThreshold = scc.GetMajorThreshold()
					valueStruct.MinorThreshold = scc.GetMinorThreshold()
					valueStruct.StatisticType = fmt.Sprint(scc.GetStatisticType())
					valueStruct.Pvalue = scc.GetPvalue()
				}(queryRegionName)
				ret := <-c
				if ret.err != nil {
					return builder.ErrorValue, fmt.Errorf("mysql_director processSub error from builder %w", ret.err)
				}
				valueStruct.Value = ret.val
				return valueStruct, nil
			} else {
				// singleThreadedDirector
				*cellCountPtr++
				scc := builder.NewTwoSampleTTestBuilder()
				_ = scc.SetKeyChain(*keychain) // ignore error
				value, err := (scc.Build(queryResult, director.statisticType, director.minorThreshold, director.majorThreshold))
				// build the value structure for this cell
				valueStruct.Path = scc.GetPath()
				valueStruct.GoodnessPolarity = scc.GetGoodnessPolarity()
				valueStruct.MajorThreshold = scc.GetMajorThreshold()
				valueStruct.MinorThreshold = scc.GetMinorThreshold()
				valueStruct.StatisticType = fmt.Sprint(scc.GetStatisticType())
				valueStruct.Pvalue = scc.GetPvalue()
				valueStruct.Value = value

				// remove this leaf key from the keychain
				if len(*keychain) > 0 {
					kc := *keychain
					kc = kc[:len(kc)-1]
					*keychain = kc
				}
				if err != nil {
					return builder.ErrorValue, fmt.Errorf("mysql_director processSub error from builder %w", err)
				}
				return valueStruct, nil
			}
		}
	} else {
		// This is a branch (not a leaf) so we keep traversing until we get a leaf.
		// Uncomment this log statement for debugging
		// 	log.Printf("mysql_director processSub branch keys are %q", keys)
		// This is a branch (not a leaf) so we keep traversing until we get to a leaf and then
		// we write the scorecard data with the value from the leaf.
		// Check to see if this is a statistic elem, so we can set the statisticType
		var keys []string = getMapKeys((region).(map[string]interface{}))
		for _, elemKey := range keys {
			for _, s := range director.statistics {
				if elemKey == fmt.Sprint(s) {
					statisticType := builder.GetStatisticTpe(elemKey)
					director.statisticType = statisticType
					break
				}
			}
			*keychain = append(*keychain, elemKey)
			queryElem := queryElem.(map[string]interface{})[elemKey]
			// Here we are in a for loop of region elements. We pass each region element (sub region) to the recursive
			// call to processSub. This is the traversal.
			// Each region element is a sub region - like a variable or a stat etc.
			// We update the region with the result of the recursive call - If the sub region is a branch the effect is that we
			// assign the region ptr back to itself, but if it is a leaf element this assigns the value structure ptr
			// into the document result map. The value structure will either be a structure or it will be an integer value (-9999).
			region.(map[string]interface{})[elemKey], err = director.processSub(queryRegionName, region.(map[string]interface{})[elemKey], queryElem, cellCountPtr, keychain, dateRange)
			// remove this branch key from the keychain
			if len(*keychain) > 0 {
				kc := *keychain
				kc = kc[:len(kc)-1]
				*keychain = kc
			}

			if err != nil {
				return builder.ErrorValue, err
			}
		}
	}
	return region, nil
}

func (director *Director) CloseDB() {
	director.db.Close()
}

// build a section of a scorecard - this is a region of a block (think vertical slice on the scorecard)
func (director *Director) Run(queryRegionName string, region interface{}, queryMap map[string]interface{}, cellCountPtr *int) (interface{}, error) {
	// This is recursive. Recurse down to the cell levl then traverse back up processing
	// all the cells on the way
	// get all the statistic strings (they are the keys of the regionMap)
	director.statistics = getMapKeys((region).(map[string]interface{})) // declared at the top
	dateRange := director.dateRange
	// declare a waitgroup so that we can wait for all the stats to finish running - only use it if !singlethreaded
	// process the regionMap (all the values will be filled in)
	var keychain []string = make([]string, 0)
	keychain = append(keychain, queryRegionName)
	region, err := director.processSub(queryRegionName, region, queryMap, cellCountPtr, &keychain, dateRange)
	// don't really care what SINGLETHREADEDDIRECTOR env var is set to, just if it is set
	_, singleThreadedDirector = os.LookupEnv("SINGLETHREADEDDIRECTOR")
	if !singleThreadedDirector {
		director.wg.Wait()
	} else {
		log.Printf("mysql_director running as SINGLETHREADEDDIRECTOR")
	}
	if err != nil {
		return region, fmt.Errorf("mysql_director error in Run %w", err)
	}
	// manager will upsert the document
	return region, nil
}
