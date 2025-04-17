package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	// Write config.yaml
	content := `database_url: "postgres://u:p@h:1/db?sslmode=disable"
embedding_dim: 42
decay_base_rate: 0.12
decay_valence_weight: 0.34
decay_interval: "2h"`
	if err := os.WriteFile("config.yaml", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://u:p@h:1/db?sslmode=disable" {
		t.Errorf("DatabaseURL = %q; want %q", cfg.DatabaseURL, "postgres://u:p@h:1/db?sslmode=disable")
	}
	if cfg.EmbeddingDim != 42 {
		t.Errorf("EmbeddingDim = %d; want %d", cfg.EmbeddingDim, 42)
	}
	if cfg.DecayBaseRate != 0.12 {
		t.Errorf("DecayBaseRate = %v; want %v", cfg.DecayBaseRate, 0.12)
	}
	if cfg.DecayValenceWeight != 0.34 {
		t.Errorf("DecayValenceWeight = %v; want %v", cfg.DecayValenceWeight, 0.34)
	}
	if cfg.DecayInterval != 2*time.Hour {
		t.Errorf("DecayInterval = %v; want %v", cfg.DecayInterval, 2*time.Hour)
	}
}

func TestLoadConfigEnvOverride(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	// Write minimal config.yaml
	content := `database_url: "x"
embedding_dim: 1
decay_base_rate: 0.1
decay_valence_weight: 0.2
decay_interval: "1h"`
	if err := os.WriteFile("config.yaml", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	// Override embedding_dim via env var
	os.Setenv("EMBEDDING_DIM", "99")
	defer os.Unsetenv("EMBEDDING_DIM")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.EmbeddingDim != 99 {
		t.Errorf("EmbeddingDim = %d; want %d", cfg.EmbeddingDim, 99)
	}
}

// TestLoadConfigDotEnv verifies that values from a .env file are loaded.
func TestLoadConfigDotEnv(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	// Write .env file with override
	if err := os.WriteFile(".env", []byte("EMBEDDING_DIM=88"), 0644); err != nil {
		t.Fatalf("Write .env failed: %v", err)
	}
	// Write config.yaml with lower value
	content := `database_url: "x"
embedding_dim: 10
decay_base_rate: 0.1
decay_valence_weight: 0.2
decay_interval: "1h"`
	if err := os.WriteFile("config.yaml", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.EmbeddingDim != 88 {
		t.Errorf("EmbeddingDim = %d; want %d", cfg.EmbeddingDim, 88)
	}
}
