package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// ----

// Holds one factoid.
type Factoid struct {
	id     int
	key    string
	value  string
	nick   string
	room   string
	date   string
	locked bool
}

// Global MySQL connection handle.
var mysqlCon *sql.DB

// Global SQLite connection handle.
var sqliteCon *sql.DB

// ----

// Wrapper function that allows to panic() with a formatted string.
func varpanic(format string, args ...interface{}) {
	msg := fmt.Sprintf("ERROR: "+format+"\n", args...)
	panic(msg)
}

// ----

func main() {
	// Die with nicer error messages.
	defer func() {
		if msg := recover(); msg != nil {
			fmt.Fprintf(os.Stderr, "%v", msg)
		}
	}()

	// Command line arguments.
	var cfgptr = flag.String("c", "analtux2gonalmatrix.ini", "Config file")
	flag.Parse()

	if stat, err := os.Stat(*cfgptr); err == nil {
		if stat.IsDir() {
			varpanic("stat %v: not a file", *cfgptr)
		}
	} else {
		varpanic("%v", err)
	}
	cfgfile := *cfgptr

	// Load the config.
	fmt.Printf("Loading configfile %v: ", cfgfile)
	cfg, err := configLoad(cfgfile)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	fmt.Printf("[okay]\n")

	var mysqlDB string
	if cfg.Section("mysql").HasKey("db") {
		mysqlDB = cfg.Section("mysql").Key("db").String()
	} else {
		varpanic("missing [mysql][db] key in config")
	}
	var mysqlPassword string
	if cfg.Section("mysql").HasKey("password") {
		mysqlPassword = cfg.Section("mysql").Key("password").String()
	} else {
		varpanic("missing [mysql][password] key in config")
	}
	var mysqlPort string
	if cfg.Section("mysql").HasKey("port") {
		mysqlPort = cfg.Section("mysql").Key("port").String()
	} else {
		varpanic("missing [mysql][port] key in config")
	}
	var mysqlServer string
	if cfg.Section("mysql").HasKey("server") {
		mysqlServer = cfg.Section("mysql").Key("server").String()
	} else {
		varpanic("missing [mysql][server] key in config")
	}
	var mysqlUser string
	if cfg.Section("mysql").HasKey("user") {
		mysqlUser = cfg.Section("mysql").Key("user").String()
	} else {
		varpanic("missing [mysql][user] key in config")
	}
	var sqliteFile string
	if cfg.Section("sqlite").HasKey("file") {
		sqliteFile = cfg.Section("sqlite").Key("file").String()
	} else {
		varpanic("missing [sqlite][file] key in config")
	}

	// Connect to MySQL.
	fmt.Printf("Connecting to MySQL: ")
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", mysqlUser, mysqlPassword, mysqlServer, mysqlPort, mysqlDB)
	mysqlCon, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	defer mysqlCon.Close()
	fmt.Printf("[okay]\n")

	// Connect to SQLite.
	fmt.Printf("Connecting to SQLite: ")
	dsn = fmt.Sprintf("%v", sqliteFile)
	sqliteCon, err = sql.Open("sqlite3", dsn)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	defer sqliteCon.Close()
	fmt.Printf("[okay]\n")

	// Create SQLite DB scheme.
	fmt.Printf("Creating table 'factoids': ")
	_, err = sqliteCon.Exec("CREATE TABLE 'factoids' (id INTEGER PRIMARY KEY AUTOINCREMENT, key STRING, value STRING, nick STRING, room STRING, timestamp DATETIME);")
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	fmt.Printf("[okay]\n")

	// Query all data from MySQL...
	fmt.Printf("Getting data from MySQL: ")
	rows, err := mysqlCon.Query("SELECT * FROM factoids")
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	defer rows.Close()
	fmt.Printf("[okay]\n")

	// ...iterate over the returned rows...
	fmt.Printf("Processing data: ")
	for rows.Next() {
		// ...extract the data from the row...
		var fact Factoid
		err = rows.Scan(&fact.id, &fact.key, &fact.value, &fact.nick, &fact.room, &fact.date, &fact.locked)
		if err != nil {
			fmt.Printf("[failed]\n")
			varpanic("%v", err)
		}

		// ...parse time...
		date, err := time.Parse("2006-01-02 15:04:05", fact.date)
		if err != nil {
			date = time.Now()
		}

		// ...make sure that the key is lower case...
		// (analtux handled keys case insensitive)
		fact.key = strings.ToLower(fact.key)

		// And write it into SQlite.
		_, err = sqliteCon.Exec(
			"INSERT INTO factoids (id, key, value, nick, room, timestamp) VALUES (?, ?, ?, ?, ?, ?);",
			fact.id, fact.key, fact.value, fact.nick, fact.room, date)
		if err != nil {
			fmt.Printf("[failed]\n")
			varpanic("%v", err)
		}
	}
	fmt.Printf("[okay]\n")
}
