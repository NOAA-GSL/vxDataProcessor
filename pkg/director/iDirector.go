package director

/*
A director holds the database connection where the actual control and
experimental data resides. For our legacy apps this is the mysql database.
For couchbase apps this is the couchbase cluster.

A director also holds a scorcard query block and a scorecard result block.
Since maps are always passed by reference - these are essentially
references to sub maps in the scorecard.

Each director is controlled by a manager. A manager has as many directors
as is needed to process Each region within a scorecard block.

The director runs as many scorecard builders in Go routines as are necessary to
process every scorecard cell within the block / region that the director is assigned.
*/
import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
)

// for couchbase all these fields will be needed
// but mysql probably only user, password, and
// host (which is actually a connection string including host and port)
type DbCredentials struct {
	User       string
	Password   string
	Host       string
	Bucket     string
	Scope      string
	Collection string
}

// see https://bitfieldconsulting.com/golang/map-string-interface
// for an explanation of any (map[string]interface{} or map[string]any)
// This is a way to define a map with a non defined structure.
type ScorecardBlock map[string]any

type Director struct {
	mysqlCredentials DbCredentials
	db               *sql.DB
	queryBlock       ScorecardBlock
	resultBlock      ScorecardBlock
	dateRange        DateRange
	minorThreshold   float64
	majorThreshold   float64
	wg               *sync.WaitGroup
	statistics       []string
	statisticType    string
}

type DirectorBuilder interface {
	// datasourceName like user:password@tcp(hostname:3306)/dbname
	Run(queryRegionName string, regionMap ScorecardBlock, queryMap ScorecardBlock)
	CloseDB()
	keys(m map[string]interface{}) []string
	getMySqlConnection(mysqlCredentials DbCredentials) (*sql.DB, error)
	queryDataPreCalc(stmnt string) (queryResult builder.PreCalcRecords, err error)
	queryDataCTC(stmnt string) (queryResult builder.CTCRecords, err error)
	queryDataScalar(stmnt string) (queryResult builder.ScalarRecords, err error)
	processSub(queryRegionName string, region interface{}, queryElem interface{}, wgPtr *sync.WaitGroup, cellCountPtr *int, keychain *[]string, dateRange DateRange) (interface{}, error)
}

type DateRange struct {
	FromSecs int64
	ToSecs   int64
}

// GetDirector returns a correctly initizalized director. Callers should make sure to call Close() when they're done with the director.
func GetDirector(directorType string, mysqlCredentials DbCredentials, dateRange DateRange, minorThreshold float64, majorThreshold float64) (*Director, error) {
	if directorType == "MysqlDirector" {
		return newMySQLDirector(mysqlCredentials, dateRange, minorThreshold, majorThreshold)
	} else {
		return nil, fmt.Errorf("Director GetDirector unsupported directorType: %q", directorType)
	}
}
