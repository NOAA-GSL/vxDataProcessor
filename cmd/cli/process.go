package main

/*
process a scorecard document
*/
import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/manager"
	"github.com/joho/godotenv"
)

func main() {
	os.Exit(process())
}

func process() int {
	defer fmt.Println("Finished")
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "document_id")
		return 1
	}

	documentId := os.Args[1]
	start := time.Now()
	// load the ${HOME}/.env if it exists
	environmentFile := fmt.Sprint(os.Getenv("HOME"), "/.env")
	err := godotenv.Load(environmentFile)
	if err != nil {
		log.Printf("Couldn't load environment file: %q", environmentFile)
	}

	mngr, err := manager.GetManager("SC", documentId)
	if err != nil {
		log.Printf("manager loadEnvironmant error GetManager %q", err)
		return 2
	}
	err = mngr.Run()
	if err != nil {
		log.Printf("manager test run error %q", err)
		return 6
	}
	elapsed := time.Since(start)
	fmt.Printf("Ttook combined %s", elapsed)
	return 0
}
