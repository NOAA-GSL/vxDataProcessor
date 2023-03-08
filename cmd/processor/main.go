package main

import (
	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
)

func main() {
	router := api.SetupRouter()

	err := router.Run(":8080") // listen and serve on 0.0.0.0:8080
	if err != nil {
		panic(err)
	}
}
