package routes

import (
	"database/sql"
	"datamin/models"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(db *sql.DB){
    r := chi.NewRouter() 
    r.Use(middleware.Logger)            // logging
    r.Use(middleware.Recoverer)         // error log tree
    r.Use(middleware.RedirectSlashes)   // site.com/profile/username/ -> site.com/profile/username
    r.Use(middleware.Throttle(1))       // research
    r.Use(middleware.Timeout(30*time.Second))
    r.Use(middleware.AllowContentEncoding("gzip","deflate"))
    // r.Use(middleware.AllowContentType("application/json", "text/xml"))
    r.Get("/", homepageHandler)

    /*

    r.Route("/scrape", func(r chi.Router){
        r.Get("/", scrapeSiteHandler)
        r.Post("/", scrapeSitePostHandler)
    })

    */

    r.Get("/scrape", scrapeSiteHandler(db))
    r.Get("/scrape/{id}", scrapeSiteSingleHandler(db))
    r.Post("/scrape/{id}", scrapeSiteSinglePostHandler(db))
    r.Get("/scrape/{id}/{singleId}/screenshot", serveScreenshot(db))
    r.Get("/static/{path}", serveStaticFiles)
    r.Post("/scrape", scrapeSitePostHandler(db))
    http.ListenAndServe(":3000", r)

}

func serveStaticFiles(w http.ResponseWriter, r *http.Request){
    
    // TODO: sanitize path

    path := fmt.Sprintf("static/%s", chi.URLParam(r, "path"))
    http.ServeFile(w, r, path)

}

func serveScreenshot(db *sql.DB) http.HandlerFunc{
    return func (w http.ResponseWriter, r *http.Request){
        ssId := chi.URLParam(r, "singleId")
        screenshotName := fmt.Sprintf("%s.png", ssId)
        path := fmt.Sprintf("media/screenshots/%s", screenshotName)
        http.ServeFile(w, r, path)
    }
}

func scrapeSiteHandler(db *sql.DB) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        temp, err := template.ParseFiles("templates/scrape.html")
        if err != nil { fmt.Printf("Error in scrapeSiteHandler: %v \n", err) }
        scrapes, err := models.GetSiteScrapesByQuery(db, "SELECT * FROM site_scrape")
        if err != nil { fmt.Printf("Error in scrapeSiteHandler (2): %v \n", err) }
        scrapes = models.LoadLastSingleToSiteScrapes(db, scrapes)  
        // print("X: " , scrapes[len(scrapes)-1].ShouldAlert("14.482"))
        data := struct {
            Scrapes []models.SiteScrape
        }{
            Scrapes : scrapes,
        }
        temp.Execute(w, data) 
        if err != nil { fmt.Printf("Err on scrapeSiteHandler (3): %v", err) }
        fmt.Printf("Serving /scrape\n")
    }
}


func scrapeSiteSingleHandler(db *sql.DB) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        temp, err := template.ParseFiles("templates/scrape-single.html")
        if err != nil { fmt.Printf("Error in scrapeSiteSingleHandler (1): %v \n", err) }
        id := chi.URLParam(r, "id")
        scrape, err := models.GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape WHERE id = '%s'", id))
        if err != nil { fmt.Printf("Error in scrapeSiteSingleHandler (2): %v \n", err) }
        err = scrape.LoadAllSingles(db)
        if err != nil { fmt.Printf("Error in scrapeSiteSingleHandler (3): %v \n", err) }
        data := struct {
            Scrape models.SiteScrape
        }{
            Scrape: scrape,
        }
        temp.Execute(w, data) 
        if err != nil { fmt.Printf("Err on scrapeSiteSingleHandler (4): %v", err) }
        fmt.Printf("Serving /scrape single\n")
    }
}

func serpSingleHandler(db *sql.DB) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        temp, err := template.ParseFiles("templates/scrape.html")
        if err != nil { fmt.Printf("Error in scrapeSiteHandler: %v \n", err) }
        serps, err := models.GetSerpsByQuery(db, "SELECT * FROM google_serp WHERE serp_url = 'https://mehonal.com/'")
        if err != nil { fmt.Printf("Error in scrapeSiteHandler (2): %v \n", err) }
        data := struct {
            Serps []models.GoogleSerp
        }{
            Serps : serps,
        }
        temp.Execute(w, data) 
        if err != nil { fmt.Printf("Err on scrapeSiteHandler (3): %v", err) }
        fmt.Printf("Serving /scrape\n")
    }
}

func scrapeSitePostHandler(db *sql.DB) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        fmt.Printf("/scrape post received\n")
        if r.FormValue("scrape-site") != "" {
            name := r.FormValue("scrape-name")
            url := r.FormValue("scrape-url")
            selector := r.FormValue("scrape-selector")
            condition := r.FormValue("scrape-condition")
            contactEmail := r.FormValue("scrape-contact-email")
            interval64, err := strconv.ParseInt(r.FormValue("scrape-interval"), 10, 64)
            active := r.FormValue("scrape-active") // 1 or 0
            getOuterHTML := r.FormValue("scrape-get-outer-html") // 1 or 0
            screenshot := r.FormValue("scrape-screenshot") // 1 or 0
            activeBool := active == "1"
            getOuterHTMLBool := getOuterHTML == "1"
            screenshotBool := screenshot == "1"
            if err != nil { w.Write([]byte("Invalid Interval. Please use a valid number.")); return; }
            interval := int(interval64)

            if !models.ValidCondition(condition) { w.Write([]byte("Invalid Condition. Leave empty for no condition or use the valid syntax for adding a condition.")); return; }
            if !models.ValidEmail(contactEmail, true) { w.Write([]byte("Invalid Email. Leave empty for no email, or type a valid email address.")); return }
            scrape, err := models.GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape WHERE url = '%s' AND css_selector = '%s'", url, selector))
            if scrape.Id < 1 {
                scrape.Name = name
                scrape.Interval = interval
                scrape.Url = url
                scrape.CssSelector = selector
                scrape.Condition = condition
                scrape.ContactEmail = contactEmail
                scrape.DateAdded = time.Now()
                scrape.Active = activeBool
                scrape.Screenshot = screenshotBool
                scrape.GetOuterHTML = getOuterHTMLBool
                }
            singleId, err := scrape.IntendScrape(db)
            go scrape.RunScrape(db, singleId) 
            if err != nil { fmt.Printf("Error in /scrape %v", err); w.Write([]byte("Error.")) }
            // http.Redirect(w, r, fmt.Sprintf("/scrape?res=%s", result), 301)
        }
        http.Redirect(w, r, "/scrape", 301)
    }
}


func scrapeSiteSinglePostHandler(db *sql.DB) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        fmt.Printf("/scrape post received\n")
        if r.FormValue("edit") != "" {
            name := r.FormValue("scrape-name")
            url := r.FormValue("scrape-url")
            selector := r.FormValue("scrape-selector")
            condition := r.FormValue("scrape-condition")
            contactEmail := r.FormValue("scrape-contact-email")
            active := r.FormValue("scrape-active") // 1 or 0
            getOuterHTML := r.FormValue("scrape-get-outer-html") // 1 or 0
            screenshot := r.FormValue("scrape-screenshot") // 1 or 0
            activeBool := active == "1"
            getOuterHTMLBool := getOuterHTML == "1"
            screenshotBool := screenshot == "1"
            interval64, err := strconv.ParseInt(r.FormValue("scrape-interval"), 10, 64)
            if err != nil { w.Write([]byte("Invalid Interval. Please use a valid number.")); return; }
            interval := int(interval64)

            if !models.ValidCondition(condition) { w.Write([]byte("Invalid Condition. Leave empty for no condition or use the valid syntax for adding a condition.")); return; }
            if !models.ValidEmail(contactEmail, true) { w.Write([]byte("Invalid Email. Leave empty for no email, or type a valid email address.")); return }
            scrape, err := models.GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape WHERE url = '%s' AND css_selector = '%s'", url, selector))
            scrape.Edit(db, name, scrape.Domain, url, selector, condition, getOuterHTMLBool, screenshotBool, activeBool, interval, contactEmail)
            if err != nil { fmt.Printf("Error in /scrape %v", err); w.Write([]byte("Error.")) }
            // http.Redirect(w, r, fmt.Sprintf("/scrape?res=%s", result), 301)
        }
        if r.FormValue("run-now") != "" {
            id := r.FormValue("scrape-id")
            scrape, err := models.GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape WHERE id = '%s'", id))
            singleId, err := scrape.IntendScrape(db)
            if err != nil { fmt.Printf("Error in /scrape %v", err); w.Write([]byte("Error.")) }
            go scrape.RunScrape(db, singleId)
            if err != nil { fmt.Printf("Error in /scrape %v", err); w.Write([]byte("Error.")) }
        }
        if r.FormValue("deactivate") != "" {
            id := r.FormValue("scrape-id")
            scrape, err := models.GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape WHERE id = '%s'", id))
            err = scrape.SetActive(db, false)
            if err != nil { fmt.Printf("Error in /scrape %v", err); w.Write([]byte("Error.")) }
        }
        if r.FormValue("activate") != "" {
            id := r.FormValue("scrape-id")
            scrape, err := models.GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape WHERE id = '%s'", id))
            err = scrape.SetActive(db, true)
            if err != nil { fmt.Printf("Error in /scrape %v", err); w.Write([]byte("Error.")) }
        }
        if r.FormValue("delete") != "" {
            id := r.FormValue("scrape-id")
            scrape, err := models.GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape WHERE id = '%s'", id))
            err = scrape.DeleteFully(db)
            if err != nil { fmt.Printf("Error in /scrape %v", err); w.Write([]byte("Error.")) }
        }

        http.Redirect(w, r, "/scrape/" + chi.URLParam(r, "id"), 301)
    }
}



func homepageHandler(w http.ResponseWriter, r *http.Request){
    http.ServeFile(w, r, "templates/index.html")
}

