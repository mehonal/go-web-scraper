package scraping

import (
	"fmt"
	"strings"
	"github.com/gocolly/colly"
)

func SearchGoogle(targetDomain string, targetKeyword string, maxTries int) (found bool, position int, serpTitle string, serpUrl string, err error){
    url := generateGoogleUrl(targetKeyword, 99) 
    c := colly.NewCollector()
    c.AllowURLRevisit = true
    proxy := GetRandomProxy()
    httpProxy, err := httpifyProxy(proxy, true)
    if err != nil {  fmt.Printf("err1: %v", err); return false, -2, serpTitle, serpUrl, err }
    err = c.SetProxy(httpProxy)
    if err != nil {  fmt.Printf("err1: %v", err); return false, -2, serpTitle, serpUrl, err }
    position = 1
    print("Making Google Serp for: " + targetKeyword + "\n")
    c.OnHTML(".egMi0", func(e *colly.HTMLElement){
        serpDiv := e.DOM
        serpTitle = serpDiv.Find("h3").Text()
        serpUrl, _ =  serpDiv.Find("a[href]").Attr("href")
        if serpUrl[0:7] == "/url?q="{
            serpUrl = serpUrl[7:]
        }

        if strings.Contains(serpUrl, "&sa=U&ved="){
            serpUrl = serpUrl[0:strings.Index(serpUrl, "&sa=U&ved=")]
        }

        if (GetHostDomain(serpUrl) == targetDomain) && !found {
            print("Found target domain: " + targetDomain + " on position " + fmt.Sprintf("%v", position) + ".\n");
            found = true

        }
        position += 1
    })
    
    c.OnRequest(func(r *colly.Request){
        destUrl := r.URL
        print(fmt.Sprintf("Visiting %s\n", destUrl))

    })

    err = c.Visit(url)
    if found == true { return found, position, serpTitle, serpUrl, err }
    if err != nil {
        print(fmt.Sprintf("Error3: %s\n", err))
        if maxTries == 0 || found == true { return found, position, serpTitle, serpUrl, err }
        found, position, serpTitle, serpUrl, err = SearchGoogle(targetDomain, targetKeyword, maxTries - 1)
    }

    return found, position, serpTitle, serpUrl, err
} 

// generates a Google Search URL based on the provided params
func generateGoogleUrl(keyword string, num_of_results int) (string){
    keyword = strings.Replace(keyword, " ", "%20", -1)
    var url string = "https://www.google.com/search?q=" + keyword + "&num=" + fmt.Sprintf("%d", num_of_results) 
    return url 
}

