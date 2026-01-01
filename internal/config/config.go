package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	SourceHost string
	DumpFile   string
	OnlyUser   string
	Help       bool
	Format     string
	MySQLUser  string
	MySQLPass  string
}

// ParseFlags parses command-line flags and returns a Config
func ParseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.SourceHost, "s", "", "Source Host")
	flag.StringVar(&cfg.DumpFile, "f", "", "Dump file")
	flag.StringVar(&cfg.OnlyUser, "o", "", "Only dump the specified user")
	flag.StringVar(&cfg.Format, "format", "raw", "Output format: raw, import, pt-like")
	flag.BoolVar(&cfg.Help, "h", false, "Print help")
	flag.Parse()
	return cfg
}

// LoadMyCnf reads the ~/.my.cnf file and sets MySQL credentials in Config
func (c *Config) LoadMyCnf() error {
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("HOME environment variable not set")
	}
	filePath := home + "/.my.cnf"
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read ~/.my.cnf: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "user=") {
			c.MySQLUser = strings.TrimSpace(line[5:])
		}
		if strings.HasPrefix(line, "password=") {
			c.MySQLPass = strings.TrimSpace(line[9:])
		}
	}
	if c.MySQLUser == "" || c.MySQLPass == "" {
		return fmt.Errorf("MySQL user or password not found in ~/.my.cnf")
	}
	return nil
}

// Validate checks if required flags are set
func (c *Config) Validate() error {
	if c.SourceHost == "" || c.DumpFile == "" {
		return fmt.Errorf("source host (-s) and dump file (-f) are required")
	}
	if c.SourceHost == c.DumpFile {
		return fmt.Errorf("source host and dump file cannot be the same")
	}
	return nil
}
