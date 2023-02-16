package main

import (
	"fmt"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	"rsc.io/quote"
)

func main() {
	fmt.Println(api.TestString())
	fmt.Println(builder.TestString())
	fmt.Println(quote.Go())
}
