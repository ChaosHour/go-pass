package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadMyCnf(t *testing.T) {
	// Create a temporary .my.cnf file
	tempDir := t.TempDir()
	tempFile := tempDir + "/.my.cnf"
	content := `[client]
user=testuser
password=testpass
`
	err := os.WriteFile(tempFile, []byte(content), 0644)
	assert.NoError(t, err)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tempDir)

	cfg := &Config{}
	err = cfg.LoadMyCnf()
	assert.NoError(t, err)
	assert.Equal(t, "testuser", cfg.MySQLUser)
	assert.Equal(t, "testpass", cfg.MySQLPass)
}

func TestLoadMyCnf_NoFile(t *testing.T) {
	// Set HOME to a non-existent directory
	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", "/nonexistent")

	cfg := &Config{}
	err := cfg.LoadMyCnf()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read ~/.my.cnf")
}

func TestLoadMyCnf_MissingCredentials(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := tempDir + "/.my.cnf"
	content := `[client]
# no user or password
`
	err := os.WriteFile(tempFile, []byte(content), 0644)
	assert.NoError(t, err)

	originalHome := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", originalHome) }()
	os.Setenv("HOME", tempDir)

	cfg := &Config{}
	err = cfg.LoadMyCnf()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MySQL user or password not found")
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				SourceHost: "127.0.0.1",
				DumpFile:   "test.sql",
			},
			wantErr: false,
		},
		{
			name: "missing source host",
			config: &Config{
				DumpFile: "test.sql",
			},
			wantErr: true,
		},
		{
			name: "missing dump file",
			config: &Config{
				SourceHost: "127.0.0.1",
			},
			wantErr: true,
		},
		{
			name: "source host and dump file same",
			config: &Config{
				SourceHost: "same",
				DumpFile:   "same",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
