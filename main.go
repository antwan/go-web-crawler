package main

import (
    "fmt"
    "time"
)


// Main function

func main() {

    // Configuration
    var ignoredPatterns = &[]string{"/cdn-cgi", "/help", "/blog"}
    const domain string = "https://server.com"
    const maxDepth int = 10
    var fetcher Fetcher = HTTPFetcher{}

    // Running crawler
    start := time.Now()
    resultsStream := make(chan *Page)
    go Crawl(domain, 0, maxDepth, ignoredPatterns, fetcher, resultsStream)

    // Retrieving results as they are fetched
    resultsCount := 0
    for page := range resultsStream {
        fmt.Println(page)
        resultsCount ++
    }
    fmt.Printf(
        "\n** Finished **\n%v URLs found in %v.\n",
        resultsCount,
        time.Since(start),
    )
}
