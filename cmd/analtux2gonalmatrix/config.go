package main

import (
	"gopkg.in/ini.v1"
)

// ----

// Load and parse the given .ini file.
func configLoad(cfgfile string) (*ini.File, error) {
	cfg, err := ini.Load(cfgfile)
	return cfg, err
}
