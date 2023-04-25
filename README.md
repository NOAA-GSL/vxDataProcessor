# vxDataProcessor

## ASCEND verification team parallel data processing project

This project is initially intended for use by the MATS project's scorecard app

This project consists of a REST API and a data builder which parallelizes the calculations
of calculation_elements. A calculation_element consists of a data set and an algorithm.
The algorithm is applied against the dataset and a result is produced. The algorithm specifies
both the expected format of the data set and the result.

## Getting Started

Currently the following works:

```shell
# Run all packages in `cmd/`
go run ./cmd/...
# Run all tests with a coverage report
go test -cover ./...

# And some tooling examples

# Format code inplace, apply simplifications if possible, and show the diff
gofmt -w -s -d .
# Run static analysis
go vet ./...
# Tidy up dependencies
go mod tidy
# Build the "test" binary
go build -o /tmp/test ./cmd/test
# Run various Linters used in CI
brew install golangci-lint # If not installed already
golangci-lint run
```

There is a cmdline cli that can be used to debug the processor or run the data processor from the terminal.
To build the cmdline cli for mac use the following command...

```bash
GOOS=darwin GOARCH=amd64 go build -o bin/mac-process cmd/cli/process.go
```

To build the cli for linux use ...

```bash
GOOS=linux GOARCH=amd64 go build -o bin/mac-process cmd/cli/process.go
```

The cli is invoked with ...

```bash
bin/mac-process "SC:anonymous--submitted:20230419150943--2block:0:03/19/2023_20_00_-_04/18/2023_13_00"
```

where the parameter is the scorecard id.

To debug the scorecard in vscode you need the following entry in your .vscode/launch.json.

```json
"version": "0.2.0",
    "configurations": [
        {
            "name": "process",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/cli/process.go",
            "env": {},
            "args": ["SC:anonymous--submitted:20230419150943--2block:0:03/19/2023_20_00_-_04/18/2023_13_00"]

        }
    ]
```

The "args" value needs to be set to the id of the scorecard document that you want to build. Once you have this launch.json
and have configured an existing scorecard id in the "args" you can run the data processor from the "Run and Debug" panel by choosing "scorecard".

TIP - Once you run the score card you can subsequently run it again by clicking the F5 key.

## Integration tests

### Environment

It helps to have the following set in your vscode settings.json file.

```json
"go.testTimeout": "10m",
"go.testTags": "integration",
"go.lintOnSave": "workspace",
"go.lintTool": "golangci-lint",
"go.lintFlags": ["--fast"],
"go.buildFlags": ["-tags=integration"],
"go.testEnvVars": {
: "/Users/randy.pierce/vxDataProcessor/.env"
},
"gopls": {
"formatting.gofumpt": true
}
```

### environment file

You need an environment file like below, although you need to get the real credentials as these are fake.
In this example the DEBUG_SCORECARD_APP_URL is pointing to a local instance of the scorecard app.
If you are running and debugging the dataprocessor and not the scorecard app you probably want to point
this to the production scorecard app UR:L. If you fail to properly designate the scorecard app URL
in your environment the manger.Run function will error trying to notify the scorecard app of the
resulting status of the run.

```bash
CB_HOST=adb-cb2.gsd.esrl.noaa.gov,adb-cb3.gsd.esrl.noaa.gov,adb-cb4.gsd.esrl.noaa.gov
CB_USER=readonlyuser
CB_PASSWORD=readonlyuserpassword
CB_BUCKET=vxdata
CB_SCOPE=_default
CB_COLLECTION=METAR
MYSQL_HOST='wolphin.fsl.noaa.gov:3306'
MYSQL_USER='mysqlreadonlyuser'
MYSQL_PASSWORD='mysqlreadonlyuserpassword'
DEBUG_SCORECARD_APP_URL=http://localhost:3000
```

### running integration tests in vscode

There are quite a few integration tests in the project. Most of them are in the manager/manager_integration_test.go.
To run these you really need to set an environment like above. The PROC_ENV_PATH points to a .env file (use your path).
You can run or debug an individual test in the context of the editor while editing a test file such as
pkg/manager/manager_integration_test.go by clicking the slightly faded "run test" or "debug test" that is located just
above the test function name. You can also use the test exlporer "flask" icon to selectively run some or all of the tests.

### running integration tests from the command line

You can run tests from the command line by cd'ing into a pkg  and running go test -run ...using a command like this.

```bash
cd ...vxDataProcessor/pkg/manager
go test -timeout 10m -tags integration -run ./...
```

or

```bash
cd .../vxdataProcessor/pkg/builder
go test -run ./...
```

the command with -v gives much more output

```bash
go test -v -run ./...
```

## Disclaimer

This repository is a scientific product and is not official communication of the
National Oceanic and Atmospheric Administration, or the United States Department
of Commerce. All NOAA GitHub project code is provided on an “as is” basis and
the user assumes responsibility for its use. Any claims against the Department
of Commerce or Department of Commerce bureaus stemming from the use of this
GitHub project will be governed by all applicable Federal law. Any reference to
specific commercial products, processes, or services by service mark, trademark,
manufacturer, or otherwise, does not constitute or imply their endorsement,
recommendation or favoring by the Department of Commerce. The Department of
Commerce seal and logo, or the seal and logo of a DOC bureau, shall not be used
in any manner to imply endorsement of any commercial product or activity by DOC
or the United States Government.
