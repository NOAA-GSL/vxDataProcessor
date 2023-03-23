package main

import (
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api/jobstore"
)

func main() {
	// TODO - benchmark if it'd be better if these channels were buffered. They will block until a reciever frees up.
	jobs := make(chan jobstore.Job)
	status := make(chan jobstore.Job)
	js := jobstore.NewJobStore()

	go api.Dispatch(jobs, js)
	go api.StatusUpdater(status, js)

	// create a pool of workers
	for w := 1; w <= 5; w++ {
		proc := &api.TestProcess{}
		go api.Worker(w, proc, jobs, status)
	}

	router := api.SetupRouter(js)

	err := router.Run(":8080") // listen and serve on 0.0.0.0:8080
	if err != nil {
		panic(err)
	}
}
