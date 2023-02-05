package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Auth holds the BasicAuth credentials for the HTTP server
type Auth struct {
	User string `mapstructure:"user,omitempty"`
	Pass string `mapstructure:"pass,omitempty"`
}

// Config holds the server configuration
type Config struct {
	Auth          Auth   `mapstructure:"auth,omitempty"`
	Port          int    `mapstructure:"port,omitempty"`
	LocalHostOnly bool   `mapstructure:"localHostOnly,omitempty"`
	SecurePort    int    `mapstructure:"securePort,omitempty"`
	CertFile      string `mapstructure:"certFile,omitempty"`
	KeyFile       string `mapstructure:"keyFile,omitempty"`
	// Max requests per second per client allowed
	RequestLimit float64 `mapstructure:"requestLimit,omitempty"`
	// If you read the documentation for ScrictSlashes
	// it lets you know that it generates a 301 redirect and converts all requests to GET requests.
	// So a POST request to /route will turn into a GET to /route/ and that will cause problems.
	// So instead you can set StripSlashes that will strip the trailing slashes before routing.
	StripSlashes bool `mapstructure:"stripSlashes,omitempty"`
	PrettyJSON   bool `mapstructure:"prettyJSON,omitempty"`
	Debug        bool `mapstructure:"debug,omitempty"`
}

func ReadConfig(config *Config) error {
	notfound := viper.ConfigFileNotFoundError{}
	// Read config from file into the Config struct
	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &notfound) {
			return err
		}
	}

	if viper.Get("debug").(bool) {
		viper.Debug()
	}

	if err := viper.Unmarshal(config); err != nil {
		return err
	}

	return nil
}

func LoadEnvConfig(rootDir string) error {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".env") {
			err = gotenv.Load(path)
			if err != nil {
				return fmt.Errorf("error loading .env file %s: %w", path, err)
			}
		}
		return nil
	})
	return err
}
