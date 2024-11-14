package database 

import (
    "fmt"
    "database/sql"
    _ "modernc.org/sqlite"
)


func GetDB() (db *sql.DB, err error){
    db, err = sql.Open("sqlite", "database.sqlite3") 
    if err != nil { return nil, err }
    // db.SetMaxIdleConns(10)
    db.SetMaxOpenConns(1)
    return db, nil
}

// Structuring 
func CreateTables(db *sql.DB) (err error){
    createTableSQL := `
        CREATE TABLE IF NOT EXISTS google_serp (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            active INTEGER,
            domain TEXT,
            keyword TEXT,
            date_added TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS google_serp_single (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            google_serp_id INTEGER,
            position INTEGER,
            position_result TEXT,
            serp_title TEXT,
            serp_url TEXT,
            date_added TIMESTAMP,
            FOREIGN KEY(google_serp_id) REFERENCES google_serp(id)
        );

        CREATE TABLE IF NOT EXISTS site_scrape (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT,
            domain TEXT,
            url TEXT,
            css_selector TEXT,
            condition TEXT,
            get_outer_html INTEGER,
            screenshot INTEGER,
            active INTEGER,
            interval INTEGER,
            contact_email TEXT,
            date_added TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS site_scrape_single (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            site_scrape_id INTEGER,
            has_screenshot INTEGER,
            screenshot_name TEXT,
            found INTEGER,
            result TEXT,
            should_alert INTEGER,
            alerted INTEGER,
            date_added TIMESTAMP,
            date_completed TIMESTAMP,
            FOREIGN KEY(site_scrape_id) REFERENCES site_scrape(id) ON DELELTE CASCADE
        );

    `
    _, err = db.Exec(createTableSQL)
    if err != nil { return err }
    fmt.Println("Table made. Tada.")
    return err
}

// Init
func InitDB(db *sql.DB) (err error) {
    print("Initing db\n")
    return CreateTables(db)    
}


