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
	id    int
	key   string
	value string
	nick  string
	room  string
	date  time.Time
}

// Our sqlite instance.
var sqliteCon *sql.DB

// ----

// Deletes all info strings for a given factoid.
func sqliteFactoidForget(key string) error {
	_, err := sqliteCon.Exec("DELETE FROM factoids WHERE key = ?;", key)
	return err
}

// Deletes a single info string for the given factoid.
func sqliteFactoidForgetValue(key string, value string) error {
	_, err := sqliteCon.Exec("DELETE FROM factoids WHERE key = ? AND value = ?;", key, value)
	return err
}

// Returns a complete info string for the given key.
func sqliteFactoidGetForKey(key string) (string, error) {
	// Query...
	rows, err := sqliteCon.Query("SELECT * FROM factoids WHERE key = ?", key)
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
		err = rows.Scan(&fact.id, &fact.key, &fact.value, &fact.nick, &fact.room, &fact.date)
		if err != nil {
			return "", err
		}

		// ...format it...
		formated := fmt.Sprintf(" %v [%v, %v]", fact.value, fact.nick, fact.date.Format("02. Jan 2006"))

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
	err = rows.Scan(&fact.id, &fact.key, &fact.value, &fact.nick, &fact.room, &fact.date)
	if err != nil {
		return "", err
	}

	// ...and format it.
	formated := fmt.Sprintf("%v = %v [%v, %v]", fact.key, fact.value, fact.nick, fact.date.Format("02. Jan 2006"))
	return formated, err
}

// Saves an info string for the given key.
func sqliteFactoidSet(key string, info string, nick string, room string) error {
	_, err := sqliteCon.Exec(
		"INSERT INTO factoids (key, value, nick, room, timestamp) VALUES  (?, ?, ?, ?, strftime('%s','now'));",
		key, info, nick, room)
	return err
}

// ----

// Connects to a SQlite database.
func sqliteConnect(sqlitefile string) error {
	dsn := fmt.Sprintf("%v", sqlitefile)
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
