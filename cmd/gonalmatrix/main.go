package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// ----

// Version number.
const VERSION_MAJOR int = 0
const VERSION_MINOR int = 1
const VERSION_PATCH int = 0

// ----

// Wrapper function that allows to panic() with a formatted string.
func varpanic(format string, args ...interface{}) {
	msg := fmt.Sprintf("ERROR: "+format+"\n", args...)
	panic(msg)
}

// ----

// Register signal handlers.
func registerSignalHandlers() {
	// SIGINT, SIGTERM.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		<-sig

		fmt.Printf("Signal received, asking syncer to stop...")
		matrixStopSyncer()
	}()
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
	var cfgptr = flag.String("c", "gonalmatrix.ini", "Config file")
	var verptr = flag.Bool("v", false, "Print version number and exit")
	flag.Parse()

	if *verptr {
		fmt.Printf("gonalmatrix v%v.%v.%v\n", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
		os.Exit(0)
	}

	if stat, err := os.Stat(*cfgptr); err == nil {
		if stat.IsDir() {
			varpanic("stat %v: not a file", *cfgptr)
		}
	} else {
		varpanic("%v", err)
	}
	cfgfile := *cfgptr

	// Print startup message
	fmt.Printf("This is gonalmatrix v%v.%v.%v\n", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
	fmt.Printf("--------------------------\n")

	// Load the config
	fmt.Printf("Loading configfile %v: ", cfgfile)
	cfg, err := configLoad(cfgfile)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	fmt.Printf("[okay]\n")

	var homeserver string
	if cfg.Section("matrix").HasKey("homeserver") {
		homeserver = cfg.Section("matrix").Key("homeserver").String()
	} else {
		varpanic("missing [matrix][homeserver] key in config")
	}
	var user string
	if cfg.Section("matrix").HasKey("username") {
		user = cfg.Section("matrix").Key("username").String()
	} else {
		varpanic("missing [matrix][username] key in config")
	}
	var passwd string
	if cfg.Section("matrix").HasKey("password") {
		passwd = cfg.Section("matrix").Key("password").String()
	} else {
		varpanic("missing [matrix][password] key in config")
	}
	var loggerfile string
	if cfg.Section("logger").HasKey("file") {
		loggerfile = cfg.Section("logger").Key("file").String()
	} else {
		varpanic("missing [logger][file] key in config")
	}
	var sqlitefile string
	if cfg.Section("sqlite").HasKey("file") {
		sqlitefile = cfg.Section("sqlite").Key("file").String()
	} else {
		varpanic("missing [sqlite][file] key in config")
	}

	// Connect to the database.
	fmt.Printf("Connecting to sqlite3 database %v: ", sqlitefile)
	err = sqliteConnect(sqlitefile)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	fmt.Printf("[okay]\n")
	defer sqliteDisconnect()

	// Connect to the server...
	fmt.Printf("Connecting to %v: ", homeserver)
	err = matrixConnect(homeserver)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	fmt.Printf("[okay]\n")

	// ...authenticate...
	fmt.Printf("Authenticating as %v: ", user)
	err = matrixAuthenticate(user, passwd)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	fmt.Printf("[okay]\n")
	defer matrixDeauthenticate()

	// ...start the event syncer...
	fmt.Printf("Starting syncer: ")
	ch := matrixStartSyncer()
	fmt.Printf("[okay]\n")
	defer close(ch)

	// ...listen for signals...
	registerSignalHandlers()

	// ...setup the event logger...
	fmt.Printf("Starting event logger: ")
	loggerhandle, err := os.OpenFile(loggerfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("[failed]\n")
		varpanic("%v", err)
	}
	matrixSetupLogger(loggerhandle)
	fmt.Printf("[okay]\n")
	defer loggerhandle.Close()

	// ...and wait forever for the syncer to finish.
	fmt.Printf("Waiting for events:\n")
	err = <-ch
	if err != nil {
		varpanic("%v", err)
	}
}
