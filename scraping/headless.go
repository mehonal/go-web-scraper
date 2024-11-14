package scraping

import (
    "fmt"
    "strings"
    "github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
    "datamin/config"
    "time"
)

func GenerateScreenshotName(hostDomain string, cssSelector string, includePath bool) (result string){
    screenshotName := ""
        fmt.Printf("Screenshotting page...\n")
        if includePath { screenshotName += "media/screenshots/"  }
        screenshotName = strings.Replace(hostDomain + "_" + cssSelector, ".", "_", -1)
        screenshotName = strings.Replace(screenshotName, " ", "_", -1)
        screenshotName += ".png"
    return screenshotName
}

func HeadlessScrapeSiteWithProxy(targetUrl string, cssSelector string, getOuterHTML bool, screenshotPage bool, screenshotName string) (result string, found bool, err error){
    fmt.Printf("Headless Scraping %s\n", targetUrl)
    l := launcher.New().Set("proxy-server", config.PROXY_LOCAL_URL).Set("ignore-certificate-errors").Set("ignore-certificate-errors-spki-list").Set("ignore-ssl-errors").MustLaunch()
    browser := rod.New().ControlURL(l).MustConnect()
    return headlessScrapeSite(targetUrl, cssSelector, getOuterHTML, screenshotPage, screenshotName, browser)
}

// not functional, needs to be updated similar to HeadlessScrapeSiteWithProxy
func HeadlessScrapeSiteWithoutProxy(targetURL string, cssSelector string, getOuterHTML bool, screenshotPage bool, screenshotName string) (result string, found bool, err error){
    fmt.Printf("Headless Scraping %s\n", targetURL)
    browser := rod.New().MustConnect()
    defer browser.MustClose()
    return headlessScrapeSite(targetURL, cssSelector, getOuterHTML, screenshotPage, screenshotName, browser)
}

func headlessScrapeSite(targetURL string, cssSelector string, getOuterHTML bool, screenshotPage bool, screenshotName string, browser *rod.Browser) (result string, found bool, err error){
    fmt.Printf("Connecting to %s\n", targetURL)
    page := browser.MustPage(targetURL)
    fmt.Printf("Connected to %s\n", targetURL)
    print("MustPaged.\n")
    page.MustWaitStable()
    time.Sleep(40 * time.Second)
    fmt.Printf("Waited till stable.\n")
    if screenshotPage {
        page.MustScreenshot("media/screenshots/" + screenshotName)
    }
    fmt.Printf("Selecting %v\n", cssSelector)
    el := page.MustElement(cssSelector)
    if getOuterHTML { result, err = el.HTML() } else { result, err = el.Text() }
    fmt.Printf("Found: %v\n", result)
    if err == nil { found = true }

    if err != nil { fmt.Printf("Error (headless.go:1): %s\n", err) }
    
    
    return result, found, err 
}


