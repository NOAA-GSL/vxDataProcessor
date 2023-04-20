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
