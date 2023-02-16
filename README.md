# vxGoDataProcessing

## AVID vx team parallel data processing project

### (initially for scorecards)

This project consists of a rest api and a data builder which parallel'izes the calculations
of calculation_elements. A calculation_element consists of a data set and an algorithm.
The algorithm is appliead against the dataset and a result is produced. The algorithm specifies
both the expected format of the data set and the result.
