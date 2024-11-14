package models

import (
    "fmt"
    "database/sql"
    "datamin/scraping"
    "errors"
    "time"
)
    

type GoogleSerp struct{
    Id int // id of the serp in the database (0 if not in database yet)
    Active bool // is the serp active
    Domain string // target domain
    Keyword string // target keyword
    DateAdded time.Time // date added to the database
}


func (serp *GoogleSerp) IntendSerpSingle(db *sql.DB) (err error){
    newSerpSingle := GoogleSerpSingle{GoogleSerpId: serp.Id, PositionResult: "Scheduled" }
    return AddSerpSingleToDB(db, newSerpSingle)
}

func (serp *GoogleSerp) RunSerpSingle(db *sql.DB, singleId int) (err error){
    found, position, serpTitle, serpUrl, err := scraping.SearchGoogle(serp.Domain, serp.Keyword, 99)
    if err != nil { return err } else { if !found { position = -1 } }
    serpSingle := GetSerpSingleByQuery(db, fmt.Sprintf(`SELECT * FROM google_serp_single WHERE id = %d`, singleId))
    serpSingle.Position = position
    serpSingle.SerpTitle = serpTitle
    serpSingle.SerpUrl = serpUrl
    return err    
}

// Quering
func GetAllSerps(db *sql.DB) (serp []GoogleSerp){
    getAllSerpsSQL := `SELECT * FROM google_serp`
    rows, err := db.Query(getAllSerpsSQL)
    serps := []GoogleSerp{}
    if err != nil { panic(err) }
    for rows.Next() {
        var serp GoogleSerp
        fmt.Printf("%v", rows.Scan(&serp.Id, &serp.Domain, &serp.Keyword, &serp.Active))
        serps = append(serps, serp)
    }
    return serps
}

func GetSerpsByQuery(db *sql.DB, query string) (serps []GoogleSerp, err error){
    rows, err := db.Query(query)
    if err != nil { panic(err) }
    for rows.Next() {
        var serp GoogleSerp
        fmt.Printf("%v", rows.Scan(&serp.Id, &serp.Domain, &serp.Keyword, &serp.Active))
        serps = append(serps, serp)
    }
    return serps, nil
}


func GetSerpByQuery(db *sql.DB, query string) (serp GoogleSerp, err error){
    rows, err := db.Query(query)
    if err != nil { panic(err) }
    for rows.Next() {
        fmt.Printf("%v", rows.Scan(&serp.Id, &serp.Domain, &serp.Keyword, &serp.Active))
        return serp, nil
    }
    return serp, nil
}




// Inserting
func AddSerpToDB(db *sql.DB, serp GoogleSerp) (err error){
    tx, err := db.Begin()
    insertSerpSQL := fmt.Sprintf(`
        INSERT INTO google_serp (domain, keyword, active) VALUES
    ('%s','%s' ,'%v');

    `, &serp.Domain, &serp.Keyword, &serp.Active)
    _, err = tx.Exec(insertSerpSQL)
    if err != nil { tx.Rollback(); return err }
    tx.Commit()
    return err
} 



// Deleting
func DeleteSerpFromDB(db *sql.DB, serpID int) (err error){
    if serpID < 1 { return errors.New("ID is below 1 in DeleteSerpFromDB") }
    _, err = db.Exec(`DELETE FROM google_serp WHERE id = ?`, serpID)
    return err 
} 


