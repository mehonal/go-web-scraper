package scraping

import (
    "fmt"
    "github.com/gocolly/colly"
)

// Scrapes a target URL for a give css selector, returning the first instance's text value
func ScrapeSite(targetURL string, cssSelector string, getOuterHTML bool) (result string, err error){
    fmt.Println("Scraping site " + targetURL)
    c := colly.NewCollector()
    
    visited := false
    
    c.OnRequest(func(req *colly.Request){
        fmt.Println("Sending request to " + targetURL)
    })

    c.OnResponse(func(resp *colly.Response){
        fmt.Printf("%v", string(resp.Body))
        fmt.Println("Response received from " + targetURL)
    })

    c.OnHTML(cssSelector, func(e *colly.HTMLElement){
        if visited { return }
        if getOuterHTML == true { result, err = e.DOM.Html() } else { result = e.Text }
        visited = true 
    }) 
    if err != nil { return result, err }
    err = c.Visit(targetURL)

    return result, err 
}
