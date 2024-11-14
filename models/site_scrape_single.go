package models 

import (
    "fmt"
    "database/sql"
    "time"
    "os"
    "datamin/mail"
)

type SiteScrapeSingle struct {
    Id int                  // unique id
    SiteScrapeId int        // parent object (site scrape) foreign key
    HasScreenshot bool      // screenshot resulting page
    ScreenshotName string   // screenshot name, if any
    Status bool             // 0 = Not Started Yet, 1 = Started, 2 = Finished
    Found bool              // css selector element found
    Result string           // value of css selector if found
    ShouldAlert bool        // has reached some condition that should trigger an alert
    Alerted bool            // alert was sent
    DateAdded time.Time     // date added to the database
    DateCompleted time.Time // date completed
    SiteScrape *SiteScrape
}

func (sss *SiteScrapeSingle) DeleteScreenshot() (err error){
    if sss.HasScreenshot {
        path := fmt.Sprintf("media/screenshots/%s", fmt.Sprintf("%d.png", sss.Id))
        err = os.Remove(path)
        if err != nil { return err }
    }
    return nil
}

func (sss *SiteScrapeSingle) EnsureParent(db *sql.DB) (err error) {
    if sss.SiteScrape == nil {
        ss, err := GetSiteScrapeByQuery(db, fmt.Sprintf("SELECT * from site_scrape WHERE id = %d", sss.SiteScrapeId))
        sss.SiteScrape = &ss
        return err
    }
    return nil
}

func (ss *SiteScrapeSingle) Alert(db *sql.DB) (err error) {
    print("Alerting ... \n")
    ss.EnsureParent(db)
    to := []string{ss.SiteScrape.ContactEmail}
    subject := fmt.Sprintf("Condition Established (%s) for '%s': %s", ss.SiteScrape.Condition, ss.SiteScrape.Name, ss.Result)
    message := fmt.Sprintf("This is an email to let you know that your desired outcome has come true for '%s' on the site %s, with the URL: %s .    \n\n What you wanted: %s  \n\n The new value is: %s", ss.SiteScrape.Name, ss.SiteScrape.Domain, ss.SiteScrape.Url, ss.SiteScrape.Condition, ss.Result)
    err = mail.SendEmail(to, subject, message)
    if err != nil { return err }
    _, err = db.Exec("UPDATE site_scrape_single SET alerted = '1' WHERE id = ?", ss.Id)
    if err != nil { return err }
    print("Email sent\n")
    return err
}


func (scs *SiteScrapeSingle) UpdateResult(db *sql.DB, result string, found bool, doAlert bool) (err error){
    dateCompleted := time.Now()
    _, err = db.Exec(fmt.Sprintf("UPDATE site_scrape_single SET result = '%s', found = '%v', should_alert = '%v' , date_completed = '%s' WHERE id = '%d'", result, found, doAlert, dateCompleted, scs.Id)) 
    if err != nil { return err }
    scs.Result = result
    scs.Found = found
    scs.ShouldAlert = doAlert
    return err 
}

func (scs *SiteScrapeSingle) UpdateAll(db *sql.DB, newScrape SiteScrapeSingle) (err error){
    scs.SiteScrapeId = newScrape.SiteScrapeId
    scs.Found = newScrape.Found
    scs.HasScreenshot = newScrape.HasScreenshot
    scs.Result = newScrape.Result
    scs.ScreenshotName = newScrape.ScreenshotName
    _, err = db.Exec(fmt.Sprintf("UPDATE site_scrape_single SET result = '%s' WHERE id = '%d'", newScrape.Result, scs.Id)) 
    return err 
}


func GetSiteScrapeSingleByQuery(db *sql.DB, query string) (siteScrapeSingle SiteScrapeSingle, err error){
    row := db.QueryRow(query)
    fmt.Printf("%v", row.Scan(&siteScrapeSingle.Id, &siteScrapeSingle.SiteScrapeId, &siteScrapeSingle.HasScreenshot, &siteScrapeSingle.ScreenshotName, &siteScrapeSingle.Found, &siteScrapeSingle.Result, &siteScrapeSingle.ShouldAlert, &siteScrapeSingle.Alerted, &siteScrapeSingle.DateAdded, &siteScrapeSingle.DateCompleted))
    return siteScrapeSingle, nil
}

func GetSiteScrapeSinglesByQuery(db *sql.DB, query string) (siteScrapeSingles []SiteScrapeSingle, err error){
    rows, err := db.Query(query)
    if err != nil { panic(err) }
    for rows.Next() {
        var siteScrapeSingle SiteScrapeSingle
        fmt.Printf("%v", rows.Scan(&siteScrapeSingle.Id, &siteScrapeSingle.SiteScrapeId, &siteScrapeSingle.HasScreenshot, &siteScrapeSingle.ScreenshotName, &siteScrapeSingle.Found, &siteScrapeSingle.Result, &siteScrapeSingle.ShouldAlert, &siteScrapeSingle.Alerted, &siteScrapeSingle.DateAdded, &siteScrapeSingle.DateCompleted))
        siteScrapeSingles = append(siteScrapeSingles, siteScrapeSingle)
    }
    return siteScrapeSingles, nil
}

func AddSiteScrapeSingleToDB(db *sql.DB, customSiteScrapeSingle SiteScrapeSingle) (id int, err error) {
    tx, err := db.Begin()
    insertSiteScrapeSingleSQL := fmt.Sprintf(`
        INSERT INTO site_scrape_single (site_scrape_id, has_screenshot, screenshot_name, found, result, should_alert, alerted, date_added, date_completed) VALUES
        ('%d', '%v', '%s', '%v', '%s', '%v', '%v', '%s', '%s');
    `, customSiteScrapeSingle.SiteScrapeId, customSiteScrapeSingle.HasScreenshot, customSiteScrapeSingle.ScreenshotName, customSiteScrapeSingle.Found, customSiteScrapeSingle.Result, customSiteScrapeSingle.ShouldAlert, customSiteScrapeSingle.Alerted, customSiteScrapeSingle.DateAdded, customSiteScrapeSingle.DateCompleted)
    res, err := tx.Exec(insertSiteScrapeSingleSQL)
    if err != nil { tx.Rollback(); return -1, err}
    id64, err := res.LastInsertId()
    if err != nil { tx.Rollback(); return -1, err}
    id = int(id64)
    if err != nil { tx.Rollback(); fmt.Printf("Error Adding: %v", err); return -1, err }
    tx.Commit()
    return id, err
}


