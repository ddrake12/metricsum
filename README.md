[![Go Reference](https://pkg.go.dev/badge/github.com/ddrake12/metricsum.svg)](https://pkg.go.dev/github.com/ddrake12/metricsum) ![Build Status](https://github.com/ddrake12/metricsum/actions/workflows/go.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/ddrake12/metricsum)](https://goreportcard.com/report/github.com/ddrake12/metricsum) 

# Introduction

metricsum is a Golang webserver module built with [`fasthttp`](https://github.com/valyala/fasthttp) and [`fasthttp/router`](https://github.com/fasthttp/router) that provides a simple API to track a statistic over the past hour. It is thread safe and can handle concurrent requests into the server without issue. 

# Installation

`go get -u github.com/ddrake12/metricsum`

# Running metricsum

Navigate to the `metricsum/cmd` folder and run any of the following commands:

 - `go run main.go` to run it in the current terminal.
 - `go build main.go`  to create a binary in the current directory.
 - `go install main.go` to create a binary at the appropriate `bin` folder. See this [reference](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies_) for details. 

## Using metricsum

Metric sum supports two endpoints, one to to track a value for a given key/statistic using `POST /metric/key` e.g.:

> `POST http://localhost:8080/metric/mykey { "value": 3}`

And one that returns the sum of all values for that key posted *within the last hour* using `GET /metric/key/sum`

> `GET http://localhost:8080/metric/mykey/sum`]

*Note*: `metricsum` uses an in memory data store and values will not persist through termination.


## Fully tested

The `metricsum` code is fully tested with appropriate mocks for external dependencies as well as the time dependent code to delete values after an hour. All Pull Requests will run the tests and ensure they pass. 

## Enjoy!