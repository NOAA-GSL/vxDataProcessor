package main

import (
	"fmt"

	"github.com/NOAA-GSL/vxGoDataProcessing/src/service"

	"rsc.io/quote"
)

func main() {
	fmt.Println(service.TestString())
	fmt.Println(quote.Go())
}
