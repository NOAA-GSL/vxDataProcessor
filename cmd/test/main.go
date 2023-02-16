package main

import (
	"fmt"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/api"
	"rsc.io/quote"
)

func main() {
	fmt.Println(api.TestString())
	fmt.Println(quote.Go())
}
