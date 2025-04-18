package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds configuration values for the CogMem library.
type Config struct {
	DatabaseURL        string        `mapstructure:"database_url"`
	EmbeddingDim       int           `mapstructure:"embedding_dim"`
	DecayBaseRate      float64       `mapstructure:"decay_base_rate"`
	DecayValenceWeight float64       `mapstructure:"decay_valence_weight"`
	DecayInterval      time.Duration `mapstructure:"decay_interval"`
	// Future Lua sandbox settings can be added here
}

// LoadConfig reads configuration from config.yaml, .env, and environment variables.
// Precedence: config.yaml -> .env file -> environment variables.
func LoadConfig() (*Config, error) {
	// Load .env file into environment if present (silently ignore missing .env)
	_ = godotenv.Load()

	v := viper.New()
    // Look for a file named config.yaml in the working directory or parent dirs
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    // Search in current directory
    v.AddConfigPath(".")
    // Also allow reading config from module root when tests run in subdirectories
    v.AddConfigPath("..")
    v.AddConfigPath("../..")

	// Read configuration file (optional)
	if err := v.ReadInConfig(); err != nil {
		// Ignore missing config file, but fail on other errors
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Allow environment variable overrides (e.g., DATABASE_URL)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %w", err)
	}

	return &cfg, nil
}
