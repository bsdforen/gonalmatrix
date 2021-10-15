package main

import (
	"database/sql"
	"fmt"
	"strings"
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

// Deletes all info strings for a given factoid.
func sqliteFactoidForget(key string) error {
	_, err := sqliteCon.Exec("DELETE FROM factoids WHERE factoid_key = ?;", key)
	return err
}

// Returns a complete info string for the given key.
func sqliteFactoidGetForKey(key string) (string, error) {
	// Query...
	rows, err := sqliteCon.Query("SELECT * FROM factoids WHERE factoid_key = ?", key)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// ...prepare string builder...
	var info strings.Builder
	info.WriteString(key)
	info.WriteString(" =")

	first := true
	for rows.Next() {
		// ...extract...
		var fact Factoid
		err = rows.Scan(&fact.id, &fact.key, &fact.text, &fact.nick, &fact.channel, &fact.date, &fact.locked)
		if err != nil {
			return "", err
		}

		// ...parse time...
		date, err := time.Parse("2006-01-02 15:04:05", fact.date)
		if err != nil {
			date = time.Now()
		}

		// ...format it...
		formated := fmt.Sprintf(" %v [%v, %v]", fact.text, fact.nick, date.Format("02. Jan 2006"))

		// ...print seperator if necessary...
		if first == false {
			info.WriteString(" ||")
		} else {
			first = false
		}

		// and at it to the info string.
		info.WriteString(formated)
	}

	if first == true {
		// Result was empty.
		return "Huh? No idea.", err
	} else {
		// We've got a result.
		return info.String(), err
	}
}

// Returns one random info string.
func sqliteFactoidGetRandom() (string, error) {
	// Query...
	rows, err := sqliteCon.Query("SELECT * FROM factoids ORDER BY RANDOM() LIMIT 1;")
	if err != nil {
		return "", err
	}
	defer rows.Close()
	rows.Next()

	// ...extract...
	var fact Factoid
	err = rows.Scan(&fact.id, &fact.key, &fact.text, &fact.nick, &fact.channel, &fact.date, &fact.locked)
	if err != nil {
		return "", err
	}

	// ...parse time...
	date, err := time.Parse("2006-01-02 15:04:05", fact.date)
	if err != nil {
		date = time.Now()
	}

	// ...and format it.
	formated := fmt.Sprintf("%v = %v [%v, %v]", fact.key, fact.text, fact.nick, date.Format("02. Jan 2006"))
	return formated, err
}

// Saves an info string for the given key.
func sqliteFactoidSet(key string, info string, author string, room string) error {
	_, err := sqliteCon.Exec(
		"INSERT INTO factoids (factoid_key, factoid_value, factoid_author, factoid_channel, factoid_timestamp, factoid_locked) VALUES  (?, ?, ?, ?, strftime('%Y-%m-%d %H:%M:%S','now'), '0');",
		key, info, author, room)
	return err
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
