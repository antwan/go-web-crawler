package main

import (
    "testing"
    "reflect"
    "sync"
)

var noFilter = &[]string{}


// 1/ Test a simple page is fetched and parsed
func testSimpleLinks(t *testing.T){

    // GIVEN: A simple website contains one page linked to 2 others
    var fetcher = MockedFetcher{
        "https://domain.com/": &MockedPage{
            "The Domain website",
            []string{
                "/a/",
                "/b/",
            },
        },
        "https://domain.com/a/": &MockedPage{
            "Page A - The Domain website",
            nil,
        },
        "https://domain.com/b/": &MockedPage{
            "Page B - The Domain website",
            nil,
        },
    }

    // WHEN Crawling the website
    resultsStream := make(chan *Page)
    results := make(map[string]string)
    go Crawl("https://domain.com/", 0, 0, noFilter, fetcher, resultsStream)
    for page := range resultsStream {
        results[page.Url] = page.Title
    }

    // THEN 3 pages are crawled
    expected := map[string]string{
        "https://domain.com/": "The Domain website",
        "https://domain.com/a/": "Page A - The Domain website",
        "https://domain.com/b/": "Page B - The Domain website",
    }
    if !reflect.DeepEqual(results, expected) {
        t.Errorf("Test failed.\nExpected: \n%s\nGot:\n%s\n", expected, results)
    }
}


// 2/ Test a simple page is not fetched twice if relinked
func testRevisitedLinks(t *testing.T){

    // GIVEN: A simple website contains one page linked to 2 others, and the two other contains previous links
    var fetcher = MockedFetcher{
        "https://domain.com/": &MockedPage{
            "The Domain website",
            []string{
                "/a/",
                "/b/",
            },
        },
        "https://domain.com/a/": &MockedPage{
            "Page A - The Domain website",
            []string{
                "/",
                "/b/",
            },
        },
        "https://domain.com/b/": &MockedPage{
            "Page B - The Domain website",
            []string{
                "/",
            },
        },
    }

    // WHEN Crawling the website
    resultsStream := make(chan *Page)
    results := make(map[string]string)
    History = &sync.Map{}
    go Crawl("https://domain.com/", 0, 0, noFilter, fetcher, resultsStream)
    for page := range resultsStream {
        results[page.Url] = page.Title
    }

    // THEN 3 pages are crawled, only once each
    expected := map[string]string{
        "https://domain.com/": "The Domain website",
        "https://domain.com/a/": "Page A - The Domain website",
        "https://domain.com/b/": "Page B - The Domain website",
    }
    if !reflect.DeepEqual(results, expected) {
        t.Errorf("Test failed.\nExpected: \n%s\nGot:\n%s\n", expected, results)
    }
}


// 3/ Test pages are crawled with absolute or relative links
func testAbsoluteLinks(t *testing.T){

    // GIVEN: A simple website contains one page linked to 2 others, one being an absolute path, the other relative
    var fetcher = MockedFetcher{
        "https://domain.com/": &MockedPage{
            "The Domain website",
            []string{
                "/a/",
                "https://domain.com/b/",
            },
        },
        "https://domain.com/a/": &MockedPage{
            "Page A - The Domain website",
            nil,
        },
        "https://domain.com/b/": &MockedPage{
            "Page B - The Domain website",
            nil,
        },
    }

    // WHEN Crawling the website
    resultsStream := make(chan *Page)
    results := make(map[string]string)
    History = &sync.Map{}
    go Crawl("https://domain.com/", 0, 0, noFilter, fetcher, resultsStream)
    for page := range resultsStream {
        results[page.Url] = page.Title
    }

    // THEN the 3 pages are crawled correctly
    expected := map[string]string{
        "https://domain.com/": "The Domain website",
        "https://domain.com/a/": "Page A - The Domain website",
        "https://domain.com/b/": "Page B - The Domain website",
    }
    if !reflect.DeepEqual(results, expected) {
        t.Errorf("Test failed.\nExpected: \n%s\nGot:\n%s\n", expected, results)
    }
}


// 4/ Test pages are not crawled if from a different domain
func testExternalLinks(t *testing.T){

    // GIVEN: A simple website contains one page linked to a same-site page, and one external link
    var fetcher = MockedFetcher{
        "https://domain.com/": &MockedPage{
            "The Domain website",
            []string{
                "/a/",
                "https://otherdomain.com/b/",
            },
        },
        "https://domain.com/a/": &MockedPage{
            "Page A - The Domain website",
            nil,
        },
    }

    // WHEN Crawling the website
    resultsStream := make(chan *Page)
    results := make(map[string]string)
    History = &sync.Map{}
    go Crawl("https://domain.com/", 0, 0, noFilter, fetcher, resultsStream)
    for page := range resultsStream {
        results[page.Url] = page.Title
    }

    // THEN Onle two pages are crawled
    expected := map[string]string{
        "https://domain.com/": "The Domain website",
        "https://domain.com/a/": "Page A - The Domain website",
    }
    if !reflect.DeepEqual(results, expected) {
        t.Errorf("Test failed.\nExpected: \n%s\nGot:\n%s\n", expected, results)
    }
}


// 5/ Test pages are not crawled if exceeding max depth
func testMaxDepth(t *testing.T){

    // GIVEN: A simple website contains several pages one linked to a different other
    var fetcher = MockedFetcher{
        "https://domain.com/": &MockedPage{
            "The Domain website",
            []string{
                "/a/",
            },
        },
        "https://domain.com/a/": &MockedPage{
            "Page A - The Domain website",
            []string{
                "/b/",
            },
        },
        "https://domain.com/b/": &MockedPage{
            "Page B - The Domain website",
            []string{
                "/c/",
            },
        },
        "https://domain.com/c/": &MockedPage{
            "Page C - The Domain website",
            []string{
                "/e/",
            },
        },
        "https://domain.com/d/": &MockedPage{
            "Page D - The Domain website",
            nil,
        },
    }

    // WHEN Crawling the website with max-depth = 4
    resultsStream := make(chan *Page)
    results := make(map[string]string)
    History = &sync.Map{}
    go Crawl("https://domain.com/", 0, 4, noFilter, fetcher, resultsStream)
    for page := range resultsStream {
        results[page.Url] = page.Title
    }

    // THEN Only 4 pages are crawled out of 4
    expected := map[string]string{
        "https://domain.com/": "The Domain website",
        "https://domain.com/a/": "Page A - The Domain website",
        "https://domain.com/b/": "Page B - The Domain website",
        "https://domain.com/c/": "Page C - The Domain website",
    }
    if !reflect.DeepEqual(results, expected) {
        t.Errorf("Test failed.\nExpected: \n%s\nGot:\n%s\n", expected, results)
    }
}


// 6/ Test pages matching a filter are excluded
func testFilters(t *testing.T){

    // GIVEN: A simple website contains several pages, some of which under a "/folder" path
    var fetcher = MockedFetcher{
        "https://domain.com/": &MockedPage{
            "The Domain website",
            []string{
                "/folder/a/",
                "/folder/b/",
                "/c/",
            },
        },
        "https://domain.com/folder/a/": &MockedPage{
            "Page A - Folder - The Domain website",
            nil,
        },
        "https://domain.com/folder/b/": &MockedPage{
            "Page B - Folder - The Domain website",
            nil,
        },
        "https://domain.com/c/": &MockedPage{
            "Page C - The Domain website",
            nil,
        },
    }

    // WHEN Crawling the website filtering on "/folder" pages
    resultsStream := make(chan *Page)
    results := make(map[string]string)
    History = &sync.Map{}
    go Crawl("https://domain.com/", 0, 0, &[]string{"/folder"}, fetcher, resultsStream)
    for page := range resultsStream {
        results[page.Url] = page.Title
    }

    // THEN only the pages out of the filter are crawled
    expected := map[string]string{
        "https://domain.com/": "The Domain website",
        "https://domain.com/c/": "Page C - The Domain website",
    }
    if !reflect.DeepEqual(results, expected) {
        t.Errorf("Test failed.\nExpected: \n%s\nGot:\n%s\n", expected, results)
    }
}
