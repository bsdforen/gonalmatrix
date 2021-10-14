package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ----

// Holds one factoid.
type Factoid struct {
	id      int
	key     string
	text    string
	nick    string
	channel string
	date    string
	locked  bool
}

// Our sqlite instance.
var sqliteCon *sql.DB

// ----

// Returns one random info string.
func sqliteFactoidGetRandom() (string, error) {
	// Query...
	rows, err := sqliteCon.Query("SELECT * FROM factoids ORDER BY RANDOM() LIMIT 1;")
	if err != nil {
		return "", err
	}
	rows.Next()

	// ...extract...
	var fact Factoid
	err = rows.Scan(&fact.id, &fact.key, &fact.text, &fact.nick, &fact.channel, &fact.date, &fact.locked)
	if err != nil {
		return "", err
	}
	rows.Close()

	// ...parse time...
	date, err := time.Parse("2006-01-02 15:04:05", fact.date)
	if err != nil {
		date = time.Now()
	}

	// ...and format it.
	formated := fmt.Sprintf("%v = %v [%v, %v]", fact.key, fact.text, fact.nick, date.Format("02. Jan 2006"))
	return formated, err
}

// ----

// Connects to a SQlite database.
func sqliteConnect(sqlitefile string) error {
	dsn := fmt.Sprintf("%v?parseTime=True", sqlitefile)
	sqlite, err := sql.Open("sqlite3", dsn)
	if err != nil {
		sqliteCon = nil
	} else {
		sqliteCon = sqlite
	}
	return err
}

// Disconnects from a SQlite database.
func sqliteDisconnect() {
	sqliteCon.Close()
}
