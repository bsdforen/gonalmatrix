package main

import (
	"flag"
	"fmt"
	"os"
)

// ----

// Version number.
const VERSION_MAJOR int = 0
const VERSION_MINOR int = 1

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
	var cfgptr = flag.String("c", "gonalmatrix.ini", "Config file")
	var verptr = flag.Bool("v", false, "Print version number")
	flag.Parse()

	if *verptr {
		fmt.Printf("gonalmatrix v%v.%v\n", VERSION_MAJOR, VERSION_MINOR)
		os.Exit(0)
	}

	if stat, err := os.Stat(*cfgptr); err == nil {
		if stat.IsDir() {
			varpanic("Not a file: %v", *cfgptr)
		}
	} else {
		varpanic("No such file: %v", *cfgptr)
	}
	cfgfile := *cfgptr
	
	// Print startup message
	fmt.Printf("This is gonalmatrix v%v.%v\n", VERSION_MAJOR, VERSION_MINOR)

	// Load the config
	cfg, err := loadConfig(cfgfile)
	if err != nil {
		varpanic("Failed to read %v", cfgfile);
	}
	homeserver := cfg.Section("global").Key("homeserver").String()
	user := cfg.Section("global").Key("username").String()
	passwd := cfg.Section("global").Key("password").String()

	// Connect to the server...
	client, err := connectMatrix(homeserver, user, passwd)
	if err != nil {
		varpanic("Couldn't connect to %v", homeserver);
	}

	// ...and start the event syncer.
	err = startSyncer(client)
	if err != nil {
		varpanic("Couldn't start syncer: %v", err);
	}
}
