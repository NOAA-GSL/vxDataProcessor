package main
/*
process a scorecard document
*/
import (
	"fmt"
	"os"
	"log"
	"time"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/manager"
)
func main() {
	os.Exit(process())
}

func process() (int){
	defer fmt.Println("Finished")
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], "environment_file document_id")
		return 1
	}
	environment_file := os.Args[1]
	documentId := os.Args[2]
	start := time.Now()
	mngr, err := manager.GetManager("SC", environment_file, documentId)
	if err != nil {
		log.Printf("manager loadEnvironmant error GetManager %q", err)
		return 2
	}
	scorecardAppUrl, err := mngr.Run()
	if err != nil {
		log.Printf("manager test run error %q", err)
		return 6
	}
	log.Printf("scorecardAppUrl is %q", scorecardAppUrl)
	elapsed := time.Since(start)
	fmt.Printf("Ttook combined %s", elapsed)
	return 0
}