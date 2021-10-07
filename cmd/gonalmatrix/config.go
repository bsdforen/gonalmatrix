package main

import (
	"gopkg.in/ini.v1"
)

// ----

func configLoad(cfgfile string) (*ini.File, error) {
	cfg, err := ini.Load(cfgfile)
	return cfg, err
}
