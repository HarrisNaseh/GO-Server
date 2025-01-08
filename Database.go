package main

import (
	"database/sql"
	"log"
	"os"
)

func dbInit() *sql.DB {
	dbName := os.Getenv("DBURL")
	db, err := sql.Open("sqlite3", dbName)

	if err != nil {
		log.Fatalf("Can not connect to database %s", err)
	}

	createString := `CREATE TABLE IF NOT EXISTS media(
        id INTEGER PRIMARY KEY NOT NULL,
        type TEXT NOT NULL,
        path TEXT NOT NULL UNIQUE,
        timestamp DATE DEFAULT CURRENT_TIMESTAMP,
        size INTEGER NOT NULL,
        mediatype TEXT NOT NULL, 
		thumbnailPath TEXT, 
		width INTEGER, 
		height INTEGER); 
		CREATE TABLE IF NOT EXISTS videoDuration( videoId INTEGER NOT NULL,
		duration INTEGER NOT NULL,
		FOREIGN KEY (videoId) REFERENCES media(id) ON DELETE CASCADE);`

	_, createErr := db.Exec(createString)

	if createErr != nil {
		log.Fatalf("Can not create table because %s", createErr)
	}

	PRAGMA_foreign_keys_String := "PRAGMA foreign_keys=ON"
	_, PRAGMAErr := db.Exec(PRAGMA_foreign_keys_String)

	if PRAGMAErr != nil {
		log.Fatalf("Can not turn on Foreign key support in database")
	}

	return db
}
