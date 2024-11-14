package models

import (
    "fmt"
    "database/sql"
    "datamin/scraping"
    "time"
)
    
type GoogleSerpSingle struct{
    Id int // id of the serp result in the database (0 if not in database yet)
    GoogleSerpId int // id of the parent in the database (0 if not in database yet)
    Position int // actual position (-2 if error, -1 if not found, 0 if not checked)
    PositionResult string // position result (-2 if error, -1 = Not Found, 0 = Scheduled, 1 or higher = Position)
    SerpTitle string // title of the serp result if found (empty if not checked)
    SerpUrl string // url of the serp result if found (empty if not checked)
    GoogleSerp GoogleSerp // parent serp
    DateAdded time.Time // date added to the database
    DateCompleted time.Time // date completed

}

func (gss *GoogleSerpSingle) CalculatePositionResult() (string){
    if gss.Position == -2 { return "Error" }
    if gss.Position == -1 { return "Not Found" }
    if gss.Position == 0 { return "Scheduled" }
    return fmt.Sprintf("%d", gss.Position)
}

func (gss *GoogleSerpSingle) LoadParent(db *sql.DB) (err error){
    gss.GoogleSerp, err = GetSerpByQuery(db, fmt.Sprintf(`SELECT * FROM google_serp WHERE id = %d`, gss.GoogleSerpId))
    return err
}

func (gss *GoogleSerpSingle) Update(db *sql.DB, position int, serpTitle string, serpUrl string) (err error){
    positionResult := gss.CalculatePositionResult()
    updateSerpSingleSQL := `UPDATE google_serp_single SET position = $1, serp_title = $2, serp_url = $3, position_result = $4, WHERE id = $5`
    _, err = db.Exec(updateSerpSingleSQL, position, serpTitle, serpUrl, positionResult, gss.Id)
    return err
}

func (gss *GoogleSerpSingle) RunSerpSingle(db *sql.DB) (err error){
    gss.LoadParent(db)
    found, position, serpTitle, serpUrl, err := scraping.SearchGoogle(gss.GoogleSerp.Domain, gss.GoogleSerp.Keyword, 99)
    if err != nil { return err } else { if !found { position = -2 } }
    gss.Position = position
    gss.SerpTitle = serpTitle
    gss.SerpUrl = serpUrl
    return AddSerpSingleToDB(db, *gss)
}

func AddSerpSingleToDB(db *sql.DB, serpSingle GoogleSerpSingle) (err error){
    addSerpSingleSQL := `INSERT INTO google_serp_single (google_serp_id, position, serp_title, serp_url) VALUES ($1, $2, $3, $4)`
    _, err = db.Exec(addSerpSingleSQL, serpSingle.GoogleSerpId, serpSingle.Position, serpSingle.SerpTitle, serpSingle.SerpUrl)
    return err
}

func GetSerpSingleByQuery(db *sql.DB, query string) (serpSingle GoogleSerpSingle){
    row := db.QueryRow(query)
    row.Scan(&serpSingle.Id, &serpSingle.GoogleSerpId, &serpSingle.Position, &serpSingle.SerpTitle, &serpSingle.SerpUrl)
    return serpSingle
}

func GetSerpSinglesByQuery(db *sql.DB, query string) (serpSingles []GoogleSerpSingle){
    rows, err := db.Query(query)
    if err != nil { panic(err) }
    for rows.Next() {
        var serpSingle GoogleSerpSingle
        rows.Scan(&serpSingle.Id, &serpSingle.GoogleSerpId, &serpSingle.Position, &serpSingle.SerpTitle, &serpSingle.SerpUrl)
        serpSingles = append(serpSingles, serpSingle)
    }
    return serpSingles
}
