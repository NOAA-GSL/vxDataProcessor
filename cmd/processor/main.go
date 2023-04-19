package main

import (
	"log"
	"os"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/manager"
	"github.com/joho/godotenv"
)

// ProcessorFactory is a wrapper function to satisfy the requirements of
// Worker and to keep the api package ignorant of the manager package
func processorFactory(docID string) (api.Processor, error) {
	return manager.GetManager(docID)
}

func main() {
	environmentFile, set := os.LookupEnv("PROC_ENV_PATH")
	if !set {
		err := godotenv.Load() // Loads from "$(pwd)/.env"
		if err != nil {
			log.Printf("Info - Unable to load .env file - %v", err)
		}
	} else {
		err := godotenv.Load(environmentFile) // Loads from whatever PROC_ENV_PATH has been set to
		if err != nil {
			log.Printf("Error - Couldn't load requested environment file at %q, error: %v", environmentFile, err)
			return
		}
	}

	// TODO - benchmark if it'd be better if these channels were buffered. They will block until a receiver frees up.
	jobs := make(chan jobstore.Job)
	status := make(chan jobstore.Job)
	js := jobstore.NewJobStore()

	go api.Dispatch(jobs, js)
	go api.StatusUpdater(status, js)

	// create a pool of workers
	for w := 1; w <= 5; w++ {
		go api.Worker(w, processorFactory, jobs, status)
	}

	router := api.SetupRouter(js)

	err := router.Run(":8080") // listen and serve on 0.0.0.0:8080
	if err != nil {
		panic(err)
	}
}
