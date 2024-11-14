package scraping

import (
    "fmt"
    "net/http"
    "io"
    "os"
    "strings"
    "math/rand"
    "datamin/config" // for PROXIES_DOWNLOAD_URL
    // "github.com/gocolly/colly"
    // "datamin/models"
)

func downloadProxiesFile() (string){
    resp, err := http.Get(config.PROXIES_DOWNLOAD_URL)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    proxiesStr := strings.Replace(string(body), "\r", "", -1)
    err = os.WriteFile("proxies.txt", []byte(proxiesStr), 0644)
    return proxiesStr
}

func formatProxiesFile() (string){
    proxiesStr, err := os.ReadFile("proxies.txt")
    if err != nil {
        panic(err)
    }
    proxies := strings.Split(string(proxiesStr), "\n")
    var formattedProxies string
    for _, proxy := range proxies {
        proxyParts := strings.Split(proxy, ":")
        if len(proxyParts) != 4 {
            continue
        }
        ip := proxyParts[0]
        port := proxyParts[1]
        username := proxyParts[2]
        password := proxyParts[3] 
        formattedProxies += fmt.Sprintf("http://%s:%s@%s:%s\n", username, password, ip, port)
    }
    err = os.WriteFile("mubeng_proxies.txt", []byte(formattedProxies), 0644)
    if err != nil {
        panic(err)
    }
    return formattedProxies
}

func GetProxies() ([]string){
    if _, err := os.Stat("proxies.txt"); os.IsNotExist(err) {
        print("Proxies file not found, downloading.")
        downloadProxiesFile()
        formatProxiesFile()
    }
    print("Reading proxies file.\n")
    proxiesStr, err := os.ReadFile("proxies.txt")
    if err != nil {
        panic(err)
    }
    proxies := strings.Split(string(proxiesStr), "\n")
    return proxies
}

func GetRandomProxy() (string){
    proxies := GetProxies()
    num := rand.Intn(len(proxies))
    return proxies[num]
}

func GetHostDomain(url string) (string){
    if strings.Contains(url, "https://") {
        url = strings.Replace(url, "https://", "", -1)    
    }
    if strings.Contains(url, "http://") {
        url = strings.Replace(url, "http://", "", -1)
    }
    if strings.Contains(url, "www.") {
        url = strings.Replace(url, "www.", "", -1)
    }
    url = strings.Split(url, "/")[0]
    url = strings.ToLower(url)
    return url
}

func httpifyProxy(proxy string, printProxy bool) (result string, err error){
    proxyParts := strings.Split(proxy, ":")
    if len(proxyParts) != 4 {
        return "", fmt.Errorf(fmt.Sprintf("Proxy has more than 4 parts: %s", proxy))
    }
    ip := proxyParts[0]
    port := proxyParts[1]
    username := proxyParts[2]
    password := proxyParts[3] 
    if printProxy {
        print(fmt.Sprintf("Proxy: %s:%s\n", ip, port))
    }
    return fmt.Sprintf("http://%s:%s@%s:%s", username, password, ip, port), nil
}

