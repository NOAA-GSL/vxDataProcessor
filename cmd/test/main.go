package main

import (
	"fmt"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder_test"
	"rsc.io/quote"
)

func main() {
	fmt.Println(api.TestString())
	fmt.Println(builder_test.TestTwoSampleTTestBuilder())
	fmt.Println(quote.Go())
}
