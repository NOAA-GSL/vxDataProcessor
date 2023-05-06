package manager

/*
A Manager is the entry point for the data processing.
A Manager holds a couchbase connection, a scorecard document id,
and a scorecard (ScorecardBlock) (which is what it
retrieves from couchbase using the id)

The manager reads the scorecard document and uses
Directors in go routines to process the region / blocks of the scorecard.
Each Region within a scorecard Block is passed to a
Director. Since maps are always passed by reference
the manager avoids duplicating data. As the Directors and their
spawned Builders build the scorecard results the scorecard
in-memory document will get filled in with results.
When a director finishes the manager will upsert the document.
There may be many upserts before the documement is fully
processed.
*/
import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/director"
	"github.com/couchbase/gocb/v2"
)

type cbConnection struct {
	Cluster    *gocb.Cluster
	Bucket     *gocb.Bucket
	Scope      *gocb.Scope
	Collection *gocb.Collection
}

type Manager struct {
	mu         sync.Mutex
	documentID string
	cb         *cbConnection
}

type ManagerBuilder interface {
	Run() error
	SetStatus(status string)
	SetProcessedAt() error
	Keys(m map[string]interface{}) []string
	loadEnvironment() (mysqlCredentials, cbCredentials director.DbCredentials, err error)
	getConnection(cbCredentials director.DbCredentials) (err error)
	upsertSubDocument(path string, subDoc interface{}) error
	getSubDocument(path string, subDocPtr *interface{}) error
	getBlocks(map[string]interface{}, error)
	getQueryBlocks(map[string]interface{}, error)
	getPlotParams(map[string]interface{}, error)
	getPlotParamCurves([]map[string]interface{}, error)
	getDateRange(director.DateRange, error)
	convertStdToPercent(std string) (percent float64, err error)
	getThresholds(plotParams map[string]interface{}) (minorThreshold, majorThreshold float64, err error)
	notifyMatsRefresh(scorecardAppURL, docID string) error
	processRegion(
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
	) error
}

func GetManager(documentID string) (*Manager, error) {
	documentType := strings.Split(documentID, ":")[0]
	if documentType == "SC" {
		return newScorecardManager(documentID)
	} else if documentType == "SCTEST" {
		_, set := os.LookupEnv("PROC_TESTING_ACCEPT_SCTEST_DOCIDS") // This should only be valid when testing
		if set {
			return newScorecardManager(documentID)
		}
	}
	return nil, fmt.Errorf("Manager GetManager unsupported documentType: %q", documentType)
}
