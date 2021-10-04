package main

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// ----

func loadConfig(cfgfile string) (*ini.File, error) {
	fmt.Printf("Loading config file %v", cfgfile)
	cfg, err := ini.Load(cfgfile)
	return cfg, err
}
