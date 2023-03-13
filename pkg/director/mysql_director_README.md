# VxDataProcessor mysql_director

The Director has the following responsibilities...

1. Recieve an app URL and a pointer to an sc_row (which is a map).
2. Query the app for the mysql query template.
3. Create a query from the template by substituting the necessary varaibles into the template
(these are embedded in the scorecard row).
4. Retrieve the input data.
5. Format the input data into the proper DerivedDataElement structures for the builders.
A derived DataElement has an InputData structure for a specific cell, and a pointer to the result
structure where the cell result value is to be placed.
6. For each data element create an inputData .
7. Fire off builders in go worker routines to process all the cell DerivedDataElement structures
   1. the builder has to do these steps...
      1. Perform a statistic calculation (RMSE, BIAS, etc on the input data) and put it into DerivedDataElement.
      2. Perform time matching on the input data
      3. Compute the significance for the DerivedDataElement
      4. write the result value into the result structure. (value is a pointer)
8. Take the value from each builder and put it into the right part of the result structure.
(maybe we should just give the builder a pointer to the result location?)

## Inputs

The manager starts a director in a go routine and gives it an sc_row structure
which has an app url, and a pointer to the row that the cell is in (which has the template
 variables e.g. region, statistic, variable - they are the keys to the specific row)

``` go
struct sc_element {
    app_url string
    row_ptr *map
    result_ptr *int
}
```

### Type

The type specifies what kind of builder is required for this data set

### data set

The data set is a JSON structure ....

### Result set

The result set is a JSON structure ...

### Algorithm

The algorithm specifies what the calculation algorithm is that is to be applied to the data set to achieve the specified result.
