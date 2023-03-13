# VxDataProcessor Builder

## A builder has to do these steps

1. Perform time matching on the input data
2. Perform a statistic calculation (RMSE, BIAS, etc on the input data) and put it into DerivedDataElement.
3. Compute the significance for the DerivedDataElement
4. write the result value into the result structure. (value is a pointer)

## Inputs

### Type

The type specifies what kind of builder is required for this data set

### data set

The data set is a JSON structure ....

### Result set

The result set is a JSON structure ...

### Algorithm

The algorithm specifies what the calculation algorithm is that is to be applied to the data set to achieve the specified result.
