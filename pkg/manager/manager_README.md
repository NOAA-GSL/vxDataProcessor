# VxDataProcessor Manager

The Manager has the following responsibilities and transformations.

1. The manager will maintain a Couchbase connection.
1. The manager waits for a process_id to be put on a queue from the service. The service
will have as many managers open as go workers as needed so that it can handle multiple service
requests simultaneously. The service starts a manager in a GO worker routine and
the manager is passed the id of the corresponding scorecard document.
1. The manager will read the scorcard document associated with the id from Couchbase
and maintain it in memory on behalf of its directors.
1. The manager will start go workers (which are directors) making sure that the number of
workers (directors) does not exceed the maximum number of database connections
configured for each kind of director. For example currently most apps are legacy apps
that require a mysql database connection. If the configuration specifies 20 allowed mysql
database connections the manager will allow up to twenty workers. Each worker is a
director and each director will maintain its own database connection (e.g. mysql client).
1. The appname associated with a scorecard block tells the manager what kind of director is
needed for each scorecard block. Each block requires an associated database query template. The
manager will build a queue of sc_element structures each of which has an appname (url?),
and a pointer to the associated result section (which has the template variables e.g. region,
statistic, variable -  they are the keys to the specific row). For example
... ```results..["rows"]["Row0"]["data"]["All HRRR domain"]["Bias (Model - Obs)"][]"2m RH"][.... ]```
1. The director will query the app that is associated with the appname for the associated template
using the app rest API.
1. Each director must derive a query (making appropriate substitutions to the template) for each
cell that needs to be calculated, then query the database for the cell data, format the data
into an InputDataElement and send the data element to an appropriate builder in a go routine. The
director uses as many GO routines as necessary to derive all the cells required of it. For example,
maybe this is one director per row, and the builder parts are delineated by region and forecastlen.
1. The builder will process the data for a given cell by...
   1. Matching the data by time.
   2. Processing the data for the associated statistic (like RMSE or BIAS).
   3. Processing the pvalue statistic.
   4. Return the result to the director.
1. The builders update the in-memory scorecard directly. When enough builders finish
the director will notify the manager when a scorecard upsert is necessary.
(perhaps when each row is complete, i.e. the director dies?)
1. The manager upserts the scorecard document with the current new results. There may be many of these upserts.
1. The manager knows that the results have all been processed when the directors have all died. The
manager does a final upsert of the scorecard, provides the return status for the service call
and then it politely dies.

## Inputs

The rest API service will call the manager with a scorecard document ID.

### Type

The type (what kind of builders are needed) is determined by the manager which will queue the appropriate
type of director. It knows what type of director by the app name (which is in each scorecard block).
When we get multiple types of apps we will need different tpyes of directors.

### data set

The scorecard is an unmarshalled JSON structure ....

### Result set

The result set is a part of the scorecard structure ...

``` json
{
    "SCORECARD": {
      "dateRange": "01/30/2023 20:00 - 03/01/2023 18:00",
      "id": "SC:anonymous--1row-at-202303030114:0:01/30/2023_20_00_-_03/01/2023_18_00",
      "name": "anonymous--1row-at-202303030114",
      "plotParams": {..},
      "processedAt": 0,
      **"results": {...}**
    }
```

It can be reached like

```SCORECARD.results```

and subsets of data may be reached like

```SCORECARD.results.`rows`.Row0.data.`All HRRR domain```
