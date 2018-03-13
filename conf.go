package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var cfg = NewConfig()

type Config struct {
	Loop         bool `json:"loop"`
	Verbose      bool `json:"verbose"`
	m_ConfigFile string
	Files        Files `json:"files"`
}

func NewConfig() *Config {
	cf := &Config{
		Loop:         false,
		Verbose:      false,
		m_ConfigFile: "",
	}
	return cf
}

func (cfg *Config) ReadConfig() error {
	if cfg.m_ConfigFile == "" {
		return nil
	}
	log.Println("Read config file:", cfg.m_ConfigFile)
	b, err := ioutil.ReadFile(cfg.m_ConfigFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, cfg)

}

func IsVerbose() bool {
	return cfg.Verbose
}
