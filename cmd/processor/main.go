package main

import (
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/manager"
)

// ProcessorFactory is a wrapper function to satisfy the requirements of
// Worker and to keep the api package ignorant of the manager package
func processorFactory(processor, docID string) (api.Processor, error) {
	return manager.GetManager(processor, docID)
}

func main() {
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
