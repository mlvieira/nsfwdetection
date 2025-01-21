package config

import (
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config holds application-wide configuration settings
type Config struct {
	Server       ServerConfig       `toml:"server"`
	Redis        RedisConfig        `toml:"redis"`
	DB           DBConfig           `toml:"database"`
	FileHandling FileHandlingConfig `toml:"file_handling"`
	Model        ModelConfig        `toml:"model"`
	Security     SecurityConfig     `toml:"security"`
}

// AppConfig stores the loaded configuration
var AppConfig Config

// LoadConfig reads the TOML configuration file
func LoadConfig(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	if err = toml.Unmarshal(data, &AppConfig); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	AppConfig.FileHandling.MaxFileSizeMB = AppConfig.FileHandling.MaxFileSizeMB << 20

	if err = os.MkdirAll(AppConfig.FileHandling.TempUploadDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create temp upload directory: %v", err)
	}

	if err = os.MkdirAll(AppConfig.FileHandling.UploadDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
		return
	}
}
