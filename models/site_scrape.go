package models

import (
	"database/sql"
	"datamin/scraping"
	"fmt"
	"time"
    "strconv"
    "regexp"
)

type SiteScrape struct {
    Id int                  // unique id
    Name string             // name of the scrap
    Domain string           // target domain
    Url string              // target url
    CssSelector string      // to scrape
    Condition string        // sends alert if condition is true for a site scrape single
    GetOuterHTML bool       // gets the element rather than its content
    Screenshot bool         // screenshot resulting page
    Active bool             // css selector element found
    Interval int            // interval in minutes
    LastSingle *SiteScrapeSingle
    Singles  *[]SiteScrapeSingle
    ContactEmail string     // email that should be alerted if condition is fulfilled
    DateAdded time.Time     // date added to the database
}

// Methods 

func (ss *SiteScrape) Edit(db *sql.DB, name string, domain string, url string, cssSelector string, condition string, getOuterHTML bool, screenshot bool, active bool, interval int, contactEmail string) (err error){
    ss.Name = name
    ss.Domain = domain
    ss.Url = url
    ss.CssSelector = cssSelector
    ss.Condition = condition
    ss.GetOuterHTML = getOuterHTML
    ss.Screenshot = screenshot
    ss.Active = active
    ss.Interval = interval
    ss.ContactEmail = contactEmail
    return ss.EditInDB(db)
}

func (ss *SiteScrape) EditInDB(db *sql.DB) (err error){
    tx, err := db.Begin()
    editSiteScrapeSQL := fmt.Sprintf(`
        UPDATE site_scrape SET name = '%s', domain = '%s', url = '%s', css_selector = '%s', condition = '%s', get_outer_html = '%v', screenshot = '%v', active = '%v', interval = '%d', contact_email = '%s' WHERE id = %d;
    `, ss.Name, ss.Domain, ss.Url, ss.CssSelector, ss.Condition, ss.GetOuterHTML, ss.Screenshot, ss.Active, ss.Interval, ss.ContactEmail, ss.Id)
    _, err = tx.Exec(editSiteScrapeSQL)
    if err != nil { tx.Rollback(); return err}
    tx.Commit()
    return err
}

func (ss *SiteScrape) ShouldIntendScrape(db *sql.DB) (res bool, err error) {
    if !ss.Active || ss.Interval == 0 { return false, nil }
    _, err = ss.LoadLastSingle(db)
    if err != nil { return false, err }
    if time.Now().Sub(ss.LastSingle.DateAdded) > time.Duration(ss.Interval) * time.Minute { return true, nil }
    return false, err
}

func (ss *SiteScrape) RunScrape(db *sql.DB, singleId int) (result string, found bool, err error){
    result, found, err = scraping.HeadlessScrapeSiteWithProxy(ss.Url, ss.CssSelector, ss.GetOuterHTML, ss.Screenshot, fmt.Sprintf("%v.png", singleId))
    if err != nil { return result, found, err } // scraping must have failed, so we don't do anything with the database
    single, err := GetSiteScrapeSingleByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape_single WHERE id = %d LIMIT 1", singleId))     
    if err != nil { return result, found, err }
    doAlert := ss.ShouldAlert(result)
    err = single.UpdateResult(db, result, found, doAlert)
    if found { if doAlert { single.Alert(db) } }
    return result, found, err 
}

func (ss *SiteScrape) ShouldAlert(result string) (bool) {
    return ShouldAlert(ss.Condition, result)
}

func (ss *SiteScrape) AddToDB(db *sql.DB) (id int, err error) {
    tx, err := db.Begin()
    insertSiteScrapeSQL := fmt.Sprintf(`
        INSERT INTO site_scrape (name, domain, url, css_selector, condition, get_outer_html, screenshot, active, interval, contact_email, date_added) VALUES
        ('%s','%s', '%s', '%s', '%s', '%v', '%v', '%v', '%d', '%s', '%s') RETURNING id;
    `, ss.Name, ss.Domain, ss.Url, ss.CssSelector, ss.Condition, ss.GetOuterHTML, ss.Screenshot, ss.Active, ss.Interval, ss.ContactEmail, ss.DateAdded)
    res, err := tx.Exec(insertSiteScrapeSQL)
    if err != nil { tx.Rollback(); return -1, err}
    id64, err := res.LastInsertId()
    if err != nil { tx.Rollback(); return -1, err}
    id = int(id64)
    if err != nil { tx.Rollback(); fmt.Printf("Error Adding: %v", err); return -1, err }
    tx.Commit()
    return id, err
}


func (ss *SiteScrape) DeleteScreenshots(db *sql.DB) (err error){
    err = ss.LoadAllSingles(db)
    if err != nil { return err }
    for i := 0; i < len(*ss.Singles); i++ {
        err = (*ss.Singles)[i].DeleteScreenshot()
        if err != nil { return err }
    }
    return err
}

func (ss *SiteScrape) DeleteFromDB(db *sql.DB) (err error){
    tx, err := db.Begin()
    deleteSiteScrapeSQL := fmt.Sprintf(`
        DELETE FROM site_scrape WHERE id = %d;
    `, ss.Id)
    _, err = tx.Exec(deleteSiteScrapeSQL)
    if err != nil { tx.Rollback(); return err}
    tx.Commit()
    return err
}

func (ss *SiteScrape) DeleteFully(db *sql.DB) (err error){
    err = ss.DeleteFromDB(db)
    if err != nil { return err }
    err = ss.DeleteScreenshots(db)
    return err
}

func (ss *SiteScrape) SetActive(db *sql.DB, active bool) (err error){
    tx, err := db.Begin()
    setActiveSQL := fmt.Sprintf(`
        UPDATE site_scrape SET active = '%v' WHERE id = %d;
    `, active, ss.Id)
    _, err = tx.Exec(setActiveSQL)
    if err != nil { tx.Rollback(); return err}
    err = tx.Commit()
    if err != nil { return err }
    ss.Active = active
    return err
}

func (ss *SiteScrape) IntendScrape(db *sql.DB) ( singleId int, err error ) {
    hostDomain := scraping.GetHostDomain(ss.Url)
    print("0: OKAY\n")
    // check if site scrape exists
    id := -5
    if ss.Id < 1 {
        fmt.Printf("Site scrape not found. Creating new site scrape.\n")
        if ss.Name == "" { ss.Name = hostDomain + "_" + ss.CssSelector }
        ss.Domain = hostDomain
        id, err = ss.AddToDB(db)
        if err != nil { fmt.Printf("Error (site_scrape.go:1): %s\n", err); return -1, err }
    } else {
        id = ss.Id
        fmt.Printf("Site scrape found.\n")
    }

    scrapeSingle := SiteScrapeSingle{
        SiteScrapeId: id,
        HasScreenshot: ss.Screenshot,
        Found: false,
        Result: "In Progress..",
        DateAdded: time.Now(),
    }
    singleId, err = AddSiteScrapeSingleToDB(db, scrapeSingle)
    return singleId, err
    
}


func (ss *SiteScrape) GetLastSingle (db *sql.DB) (result SiteScrapeSingle, err error) {
    result, err = GetSiteScrapeSingleByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape_single WHERE site_scrape_id = %d ORDER BY id DESC LIMIT 1", ss.Id))
    return result, err
}


func (ss *SiteScrape) LoadLastSingle (db *sql.DB) (scrape SiteScrape, err error) {
    single, err := GetSiteScrapeSingleByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape_single WHERE site_scrape_id = %d ORDER BY id DESC LIMIT 1", ss.Id))
    ss.LastSingle = &single 
    return scrape, err
}

func (ss *SiteScrape) LoadAllSingles (db *sql.DB) (err error) {
    singles, err := GetSiteScrapeSinglesByQuery(db, fmt.Sprintf("SELECT * FROM site_scrape_single WHERE site_scrape_id = %d ORDER BY id DESC", ss.Id))
    if err != nil { return err }
    ss.Singles = &singles
    return err
}

// Functions


/* Condition Possible Values 

num:<5
num:>42.19
res:Some Text

*/

func ExtractNumbersAndDecimals(input string) string {
	re := regexp.MustCompile(`[^\d.]`)
	return re.ReplaceAllString(input, "")
}

func ValidEmail(email string, acceptEmpty bool) (bool) {
    if acceptEmpty && email == "" { return true }
    emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
} 

func ValidCondition(condition string) (bool) {
    if condition == "" { return true }
    if len(condition) < 4 { return false }
    if condition[:4] == "num:" || condition[:4] == "res:" {
        return true
    }
    return false
}

func ShouldAlert(condition string, result string) (bool){
    if condition == "" { return false }
    if condition[:3] == "num"{
        print("ok1\n")
        thresholdVal, err := strconv.ParseFloat(condition[5:], 64)
        print("ok2", thresholdVal, "\n")
        if err != nil { fmt.Printf("Error on ShouldAlert: %v\n", err); return false }
        currentRes, err := strconv.ParseFloat(ExtractNumbersAndDecimals(result), 64)
        print("ok3", currentRes, "\n")
        if err != nil { fmt.Printf("Error on ShouldAlert: %v\n", err); return false }
        op := condition[4]
        if op == '=' { return thresholdVal == currentRes }
        if op == '>' { return currentRes > thresholdVal }
        if op == '<' { return currentRes < thresholdVal }
    }
    return false
}

func LoadLastSingleToSiteScrapes(db *sql.DB, scrapes []SiteScrape) ([]SiteScrape){
    for i := 0; i < len(scrapes); i ++ {
        scrapes[i].LoadLastSingle(db)
    }
    return scrapes
}

func GetSiteScrapesByQuery(db *sql.DB, query string) (siteScrapes []SiteScrape, err error){
    rows, err := db.Query(query)
    if err != nil { panic(err) }
    for rows.Next() {
        var siteScrape SiteScrape
        fmt.Printf("%v", rows.Scan(&siteScrape.Id, &siteScrape.Name, &siteScrape.Domain, &siteScrape.Url, &siteScrape.CssSelector, &siteScrape.Condition, &siteScrape.GetOuterHTML, &siteScrape.Screenshot, &siteScrape.Active, &siteScrape.Interval, &siteScrape.ContactEmail, &siteScrape.DateAdded))

        siteScrapes = append(siteScrapes, siteScrape)
    }
    return siteScrapes, nil
}


func GetSiteScrapeByQuery(db *sql.DB, query string) (siteScrape SiteScrape, err error){
    row := db.QueryRow(query)
    fmt.Printf("%v", row.Scan(&siteScrape.Id, &siteScrape.Name, &siteScrape.Domain, &siteScrape.Url, &siteScrape.CssSelector, &siteScrape.Condition, &siteScrape.GetOuterHTML, &siteScrape.Screenshot, &siteScrape.Active, &siteScrape.Interval, &siteScrape.ContactEmail, &siteScrape.DateAdded))
    return siteScrape, nil
}

func AddSiteScrapeToDB(db *sql.DB, ss SiteScrape) (id int, err error) {
    tx, err := db.Begin()
    insertSiteScrapeSQL := fmt.Sprintf(`
        INSERT INTO site_scrape (name, domain, url, css_selector, condition, get_outer_html, screenshot, active, interval, contact_email, date_added) VALUES
        ('%s','%s', '%s', '%s', '%s', '%v', '%v', '%v', '%d', '%s', '%s') RETURNING id;
    `, ss.Name, ss.Domain, ss.Url, ss.CssSelector, ss.Condition, ss.GetOuterHTML, ss.Screenshot, ss.Active, &ss.Interval, ss.ContactEmail, ss.DateAdded)
    res, err := tx.Exec(insertSiteScrapeSQL)
    if err != nil { tx.Rollback(); return -1, err}
    id64, err := res.LastInsertId()
    if err != nil { tx.Rollback(); return -1, err}
    id = int(id64)
    if err != nil { tx.Rollback(); fmt.Printf("Error Adding: %v", err); return -1, err }
    tx.Commit()
    return id, err
}

