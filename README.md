# About

This is a personal project with the aim of exploring the capabilities of Golang in the realm of scraping - and particularly headless browser scraping.

The project utilizes:
- Web scraping through net/http and Go-colly
- Headless Web scraping via Go-rod
- Rotating proxies through Mubeng
- Web interface provided via simple html, css, js routed through Go-chi

## Pages

### Scrape Site Archive
- Includes: add new, see existing

### Scrape Site Single
- Includes: edit/delete existing, see history

### Serp Archive
- Includes: add new, see existing

### Serp Single
- Includes: edit/delete existing, see history

## To Add Later
- Bulk Actions

# Future considerations
1. Cleanse inputs; perhaps a library can be utilized to handle this
2. Security 

# How to run
## 1. Install Mubeng:

Github: https://github.com/kitabisa/mubeng

> Installation:

▶ git clone https://github.com/kitabisa/mubeng
▶ cd mubeng
▶ make build
▶ (sudo) mv ./bin/mubeng /usr/local/bin
▶ make clean

## 2. Run Mubeng:

Example usage:
` $ mubeng -f mubeng_proxies.txt -r 100 -a 127.0.0.1:9453 `

(Add -w if the proxies file will change)

## 3. Set up config/supersecret.go

Your supersecret file contains sensitive info. Here is a sample supersecret.go file:

```
package config

var PROXIES_DOWNLOAD_URL string = "https://example.com/" // returns txt file with proxies in format: IP:port:username:password

var PROXY_LOCAL_URL string = "127.0.0.1:9453"

var SMTP_SERVER string = "mail.example.com"
var SMTP_PORT int = 587
var SMTP_EMAIL string = "mail@example.com"
var SMTP_PASS string = "example-pass"
var GPT_AUTH_TOKEN string = "sk-gpt_auth_token"

```

## 4. Run main.go

Simply run main.go using `go run main.go`

Congrats! You should now be running the scraper on `localhost:3000`


---

# Potential Issues
- SSL-enforced content not visible (ie lottie animation)

---

# Tested working code

## Test Website Price

`
result, err := scraping.ScrapeSite("https://website.com/", "span.price > span.res", false)
if err != nil { fmt.Printf("Error: %v\n", err) }
if result == "" { result = "Not Found"}
fmt.Printf("Res: %v\n", result)
`

# Important Notice

 The project is publicized for educational purposes, and should not be used for misconduct. Please ensure you have the permission to scrape websites that you scrape if you choose to use the tool.
