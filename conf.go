package main

var cfg = NewConfig()

type Config struct {
	m_Loop bool
	m_Verbose bool
	m_OutputFile string
}

func NewConfig() *Config {
	cf := &Config{
		m_Loop: false,
		m_Verbose: false,
		m_OutputFile: "",
	}
	return cf
}

func (cfg *Config)ReadConfig() error {
	return nil
}

func IsVerbose() bool {
	return cfg.m_Verbose
}
