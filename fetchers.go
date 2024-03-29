package main

import (
    "io"
    "fmt"
    "net/http"
    "strings"
)

// Fetcher can return a document given a URL
type Fetcher interface {
    Fetch(url string) (body io.ReadCloser, err error)
    Close(reader io.ReadCloser)
}


// HTTPFetcher is a fetcher that returns real results using HTTP requests
type HTTPFetcher struct {}

// Fetch retrieves the URL data and returns it
func (fetcher HTTPFetcher) Fetch(url string) (io.ReadCloser, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    return resp.Body, nil
}

// Close closes the data reader
func (fetcher HTTPFetcher) Close(response io.ReadCloser) {
    if response != nil {
        response.Close()
    }
}


// MockedFetcher is a fetcher that return predefined results for testing purposes
type MockedFetcher map[string]*MockedPage

// MockedPage is the underlying structure to build MockedFetcher
type MockedPage struct {
    Title string
    Links []string
}

type stringReaderCloser struct {
    *strings.Reader
}

func (r stringReaderCloser) Close() error {
    return nil
}

// Fetch retrieves the URL data and returns it
func (fetcher MockedFetcher) Fetch(url string) (io.ReadCloser, error) {
    if page, ok := fetcher[url]; ok {
        linksChunk := ""
        for _, link := range page.Links {
            linksChunk += fmt.Sprintf("<a href=\"%s\">Some link</a>\n", link)
        }
        body := fmt.Sprintf("<html><title>%s</title>\n<body>%s</body></html>", page.Title, linksChunk)
        return stringReaderCloser{strings.NewReader(body)}, nil
    }
    return nil, fmt.Errorf("not found: %s", url)
}

// Close closes the data reader
func (fetcher MockedFetcher) Close(reader io.ReadCloser) {}
