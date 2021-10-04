package main

import (
	"flag"
	"fmt"
	"os"
)

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
	flag.Parse()

	if stat, err := os.Stat(*cfgptr); err == nil {
		if stat.IsDir() {
			varpanic("Not a file: %v", *cfgptr)
		}
	} else {
		varpanic("No such file: %v", *cfgptr)
	}
	cfgfile := *cfgptr

	// Load the config
	cfg, err := loadConfig(cfgfile)
	if err != nil {
		varpanic("Failed to read %v", cfgfile);
	}
}
