package main

import (
	"datamin/database"
	"datamin/gpt"
	"datamin/mail"
	"datamin/models"
	"datamin/routes"
	"datamin/scraping"
	"fmt"
	"github.com/robfig/cron"
	// "datamin/utils"
	// "datamin/models"
	"database/sql"
    "time"
)

func main() {
    fmt.Printf("Firing up Datamin...\n")
    db, err := database.GetDB()
    if err != nil { fmt.Printf("Error: %s\n", err) }
    defer db.Close()
    database.InitDB(db)
    setupCronJobs(db)
    routes.SetupRouter(db)
}

func runScheduledScrapes(db *sql.DB) {
    fmt.Printf("Running scheduled scrapes...\n")
    scrapes, err := models.GetSiteScrapesByQuery(db, "SELECT * FROM site_scrape WHERE active = 'true' AND interval != 0")
    fmt.Printf("Found %d scrapes\n", len(scrapes))
    if err != nil { fmt.Printf("Error during runScheduledScrapes: %s\n", err) }
    for i := 0; i < len(scrapes); i++ {
        shouldScrape, err:= scrapes[i].ShouldIntendScrape(db)
        if err != nil { fmt.Printf("Error during runScheduledScrapes: %s\n", err); continue; }
        if shouldScrape {
            fmt.Printf("Intending to scrape %s\n", scrapes[i].Name)
            singleId, err := scrapes[i].IntendScrape(db)
            if err != nil { fmt.Printf("Error during runScheduledScrapes: %s\n", err); continue; }
            scrapes[i].RunScrape(db, singleId)
            time.Sleep(10 * time.Second) // generally, it takes at least 30 seconds to scrape a site
        }
    } 

}

func setupCronJobs(db *sql.DB) {
    c := cron.New()
    c.AddFunc("@every 10s", func() { fmt.Printf("cron on the loose\n"); runScheduledScrapes(db) })
    c.Start()
} 
























// Examples

func askGptExample() {
    gpt.AskGpt("You are a mathamagician with a sense of humour. Make sure you always use a joke while answering.", "How can I find c if I know a is 3 and b is 4 in a triangle?")
}


func sendEmailEg() { 
    mail.SendEmail([]string{"somemail@domain.com"},"Test email", "this is a testing email.")
}

func scrapeSiteEg() {
    db, err := database.GetDB()
    if err != nil { fmt.Printf("Error: %s\n", err) }
    err = database.InitDB(db)
    if err != nil { fmt.Printf("Error: %s\n", err) }
    print("Headless scraping mehonal out \n") 
    // scraping.HeadlessScrapeSiteWithProxy(db, "https://mehonal.com", "h1", false, false, "127.0.0.1:9453")    
    print("Done headless scraping out\n")
}

func scrapeGoogleResultEg() {
    db, err := database.GetDB()
    if err != nil { fmt.Printf("Error: %s\n", err) }
    err = database.InitDB(db)
    if err != nil { fmt.Printf("Error: %s\n", err) }
    print("serping out \n") 
    scraping.SearchGoogle("mehonal.com", "Mehonal Niagara Web Design", 15)
    print("Done trying serp\n")
}
