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
	var configptr = flag.String("c", "gonalmatrix.ini", "Config file")
	flag.Parse()

	if stat, err := os.Stat(*configptr); err == nil {
		if stat.IsDir() {
			varpanic("Not a file: %v", *configptr)
		}
	} else {
		varpanic("No such file: %v", *configptr)
	}
	configfile := *configptr
}
