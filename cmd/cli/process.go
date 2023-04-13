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

	documentID := os.Args[1]
	start := time.Now()
	environmentFile, set := os.LookupEnv("PROC_ENV_PATH")
	if !set {
		err := godotenv.Load() // Loads from "$(pwd)/.env"
		if err != nil {
			log.Printf("Couldn't load environment file: %q", environmentFile)
		}
	} else {
		err := godotenv.Load(environmentFile) // Loads from whatever PROC_ENV_PATH has been set to
		if err != nil {
			log.Printf("Couldn't load environment file: %q", environmentFile)
			return 7
		}
	}

	mngr, err := manager.GetManager("SC", documentID)
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
	fmt.Printf("Took combined %s", elapsed)
	return 0
}
