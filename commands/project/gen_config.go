package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DB struct {
		Postgres struct {
			URI                   string `json:"uri"`
			MaxOpenConnections    int    `json:"max_connections"`
			MaxIdleConnections    int    `json:"max_idle_connections"`
			MaxConnectionIdleTime int    `json:"max_connection_idle_time"`
			MaxConnectionLifetime int    `json:"max_connection_lifetime"`
		} `json:"postgres"`
	} `json:"db"`
}

func WriteConfig(projectPath string) error {
	// write config struct to config.go
	configContent := `
  package config

  type Config struct {
      DB struct {
          Postgres struct {
              URI                   string ` + "`json:\"uri\"`" + `
              MaxOpenConnections    int    ` + "`json:\"max_connections\"`" + `
              MaxIdleConnections    int    ` + "`json:\"max_idle_connections\"`" + `
              MaxConnectionIdleTime int    ` + "`json:\"max_connection_idle_time\"`" + `
              MaxConnectionLifetime int    ` + "`json:\"max_connection_lifetime\"`" + `
          } ` + "`json:\"postgres\"`" + `
      } ` + "`json:\"db\"`" + `
  }`

	configPath := filepath.Join(projectPath, "internal", "infrastructure", "config", "config.go")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config.go: %v", err)
	}

	// default config template
	defaultConfig := Config{}
	defaultConfig.DB.Postgres = struct {
		URI                   string `json:"uri"`
		MaxOpenConnections    int    `json:"max_connections"`
		MaxIdleConnections    int    `json:"max_idle_connections"`
		MaxConnectionIdleTime int    `json:"max_connection_idle_time"`
		MaxConnectionLifetime int    `json:"max_connection_lifetime"`
	}{
		URI:                   "postgres://username:password@127.0.0.1:5432/dbname",
		MaxOpenConnections:    25,
		MaxIdleConnections:    25,
		MaxConnectionIdleTime: 300,
		MaxConnectionLifetime: 3600,
	}

	// write to existing json config files
	envFiles := map[string]string{
		"dev":     filepath.Join(projectPath, "configs", "dev.json"),
		"staging": filepath.Join(projectPath, "configs", "stg.json"),
		"prod":    filepath.Join(projectPath, "configs", "prod.json"),
	}

	for _, path := range envFiles {
		configData, err := json.MarshalIndent(defaultConfig, "", "    ")
		if err != nil {
			return fmt.Errorf("error marshaling config: %v", err)
		}

		if err := os.WriteFile(path, configData, 0644); err != nil {
			return fmt.Errorf("error writing config to %s: %v", path, err)
		}
	}

	return nil
}

func WriteConfigLoader(projectPath string) error {
	configLoaderContent := `// internal/infrastructure/config/load.go
package config

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
)

func LoadConfig(env string) (*Config, error) {
    if env == "" {
        env = "dev"
    }

    // get project root by walking up from this file's location
    _, filename, _, _ := runtime.Caller(0)
    projectRoot := filepath.Join(filepath.Dir(filename), "../../..")

    configPath := filepath.Join(projectRoot, "configs", fmt.Sprintf("%s.json", env))
    if _, err := os.Stat(configPath); err != nil {
        return nil, fmt.Errorf("config file not found for environment %s: %v", env, err)
    }

    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("error reading config file: %v", err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("error parsing config file: %v", err)
    }

    return &cfg, nil
}

func GetEnv() string {
    env := os.Getenv("ENV")
    if env == "" {
        return "dev"
    }
    return env
}`

	configLoaderPath := filepath.Join(projectPath, "internal", "infrastructure", "config", "load.go")

	if err := os.WriteFile(configLoaderPath, []byte(configLoaderContent), 0644); err != nil {
		return fmt.Errorf("failed to write load.go: %v", err)
	}

	return nil
}
