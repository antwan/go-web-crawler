package main

import (
    "fmt"
    "sync"
    "io"
    "net/url"
    "golang.org/x/net/html"
    "strings"
)


// SafeCounter holds a counter that's safe to access concurrently
// (We can use alternatively sync/atomic)

type SafeCounter struct {
    val uint
    mux sync.Mutex
}

func (counter *SafeCounter) Increase() {
    counter.mux.Lock()
    counter.val += 1
    counter.mux.Unlock()
}
func (counter *SafeCounter) Decrease() {
    counter.mux.Lock()
    counter.val -= 1
    counter.mux.Unlock()
}
func (counter *SafeCounter) IsEmpty() bool {
    counter.mux.Lock()
    defer counter.mux.Unlock()
    return counter.val == 0
}


// Page is a structure that contains a title and sublinks

type Page struct {
    Url string
    Title string
    Error error
    SubLinks []string
    Depth int
}

func (page *Page) String() string {
    depthString := ""
    if page.Depth >= 1 {
        for i := 0; i < page.Depth; i++ {
            depthString += "  "
        }
        depthString += "\\_ "
    }
    if page.Error != nil {
        return fmt.Sprintf("%s%s /!\\ Error: %s", depthString, page.Url, page.Error)
    }
    return fmt.Sprintf("%s%s [%s] (with %d sublinks)", depthString, page.Url, page.Title, len(page.SubLinks))
}


// Processes a link and stores it to the map if necessary

func ProcessLink(uri, base string, ignoredPatterns *[]string, store *map[string]bool) {

    parsedUri, err := url.Parse(uri)
    if err != nil {
        return
    }
    parsedBase, _ := url.Parse(base)
    baseHost := parsedBase.Host

    // Converts relative URLs to absolute
    parsedUri = parsedBase.ResolveReference(parsedUri)


    // Ignores URL fragments
    parsedUri.Fragment = ""

    // Processes only the links to same domain
    if (parsedUri.Host != baseHost) {
        return
    }

    // Processes only the new links
    _, existing := (*store)[parsedUri.String()]
    if existing {
        return
    }

    // Filters out URLs that are excluded
    filtered := false
    for _, prefix := range *ignoredPatterns {
        if strings.HasPrefix(parsedUri.Path, prefix) {
            filtered = true
            break
        }
    }
    if filtered {
        return
    }

    // Saves link in the map
    (*store)[parsedUri.String()] = true
}


// Parses a HTML document and returns title and list of links
// Silently finishes if an error is met (e.g when parsing invalid HTML document)

func ParseDocument(htmlBody io.Reader, parentUri string, ignoredPatterns *[]string) (string, []string) {

    tokenizer := html.NewTokenizer(htmlBody)
    var title string
    var links map[string]bool = make(map[string]bool)
    finished := false

    // Looping through all page tokens
    for !finished {
        token := tokenizer.Next()
        switch {
        case token == html.ErrorToken:
            finished = true
        case token == html.StartTagToken:
            tag := tokenizer.Token()
            if tag.Data == "a" {
                for _, attr := range tag.Attr {
                    if attr.Key == "href" {
                        // Found a valid <a href=""> tag
                        ProcessLink(attr.Val, parentUri, ignoredPatterns, &links)
                    }
                }
            }
            // Found a valid <title> tag
            if tag.Data == "title" {
                tokenizer.Next()
                title = strings.TrimSpace(tokenizer.Token().Data)
            }
        }
    }

    // Converts map to slice for return
    subLinks := make([]string, 0, len(links))
    for u := range links {
        subLinks = append(subLinks, u)
    }
    return title, subLinks
}


// Crawl fetches URL and crawl recursively nested urls, returning results in stream

func Crawl(uri string, depth int, maxDepth int, ignoredPatterns *[]string, fetcher Fetcher, resultsStream chan *Page) {

    // Initialize crawler and prepare for closure once last job is complete
    if depth == 0 {
        jobsCounter.Increase()
    }
    defer func() {
        jobsCounter.Decrease()
        if jobsCounter.IsEmpty() {
            close(resultsStream)
        }
    }()

    // MaxDepth prevents recursive call to loop infinitely (if set)
    if maxDepth > 0 && depth >= maxDepth {
        return
    }

    // Fetches page from fetcher
    resp, err := fetcher.Fetch(uri)
        defer fetcher.Close(resp)
    History.Store(uri, true)

    if err != nil {
        resultsStream <- &Page{Url: uri, Error: err, Depth: depth}
        return
    } else {
        // Valid page is found, parsing it
        title, subLinks := ParseDocument(resp, uri, ignoredPatterns)
        resultsStream <- &Page{Url:uri, Title:title, SubLinks:subLinks, Depth:depth}

        // Crawling nested links if they are not visited yet
        for _, u := range subLinks {
            _, existing := History.LoadOrStore(u, true)
            if !existing {
                // jobCounter is increased before goroutine call, to avoid parent thread to close channel
                // before child is started
                jobsCounter.Increase()
                go Crawl(u, depth + 1, maxDepth, ignoredPatterns, fetcher, resultsStream)
            }
        }
        return
    }
    return

}


// Caches to control routines spawns

var History = &sync.Map{}
var jobsCounter = &SafeCounter{}
