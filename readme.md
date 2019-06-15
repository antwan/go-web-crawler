# Go crawler

A simple Web crawler written in go.
This features possiblity to exclude some URLs, and limit recursion depth.

## Instructions

* Download package locally
* Run package using `go run .`
* Edit configuration in `main.go` to change settings.

## Design notes

* Crawler is using routines for efficiency. Several implementations are possible to manage all edge-cases efficiently (e.g fifo stack, limited workers, etc). For simplicity, this is just using one goroutine per request.
* Few tests are provided. To run them use `go test` from package directory.

## Extensions

The implementation is kept simple for the purpose of this exercice, and several features are not implemented. Possible extensions include:
* Timeout for crawler
* More sophisticated storage structure to order URLs by hierarchy
* Argument parsers for runner.
* Improved error handling for different HTTP statuses (404, etc)
* etc.

