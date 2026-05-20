package bootstrap

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
}

type HTTPConfig struct {
	Addr string
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func LoadConfig() (*Config, error) {
	fileEnv, err := loadFileEnv()
	if err != nil {
		return nil, err
	}

	maxOpenConns, err := loadInt(fileEnv, "DB_MAX_OPEN_CONNS", 20)
	if err != nil {
		return nil, err
	}
	maxIdleConns, err := loadInt(fileEnv, "DB_MAX_IDLE_CONNS", 10)
	if err != nil {
		return nil, err
	}
	connMaxLifetime, err := loadDuration(fileEnv, "DB_CONN_MAX_LIFETIME", time.Hour)
	if err != nil {
		return nil, err
	}

	return &Config{
		HTTP: HTTPConfig{Addr: loadString(fileEnv, "HTTP_ADDR", ":8080")},
		DB: DBConfig{
			DSN:             loadString(fileEnv, "DB_DSN", "root:password@tcp(127.0.0.1:3306)/anserflow?charset=utf8mb4&parseTime=True&loc=Local"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
		},
	}, nil
}

func loadFileEnv() (map[string]string, error) {
	root, err := findMonorepoRoot()
	if err != nil {
		return nil, err
	}

	envName := os.Getenv("NODE_ENV")
	if envName == "" {
		envName = "development"
	}

	merged := make(map[string]string)
	paths := []string{
		filepath.Join(root, ".env"),
		filepath.Join(root, ".env."+envName),
		filepath.Join(root, ".env.local"),
	}

	for _, path := range paths {
		values, err := loadEnvFile(path)
		if err != nil {
			return nil, err
		}
		for key, value := range values {
			merged[key] = value
		}
	}

	return merged, nil
}

func findMonorepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return findMonorepoRootFrom(dir)
}

func findMonorepoRootFrom(dir string) (string, error) {
	for {
		gitPath := filepath.Join(dir, ".git")
		_, err := os.Stat(gitPath)
		if err == nil {
			return dir, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat %s: %w", gitPath, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("find monorepo root: .git not found")
		}
		dir = parent
	}
}

func loadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	values, err := parseDotEnv(file)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return values, nil
}

func parseDotEnv(r io.Reader) (map[string]string, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: missing '='", lineNo)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNo)
		}

		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			if value[0] == '\'' && value[len(value)-1] == '\'' {
				value = value[1 : len(value)-1]
			}
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan .env: %w", err)
	}
	return values, nil
}

func loadString(fileEnv map[string]string, key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	value, ok = fileEnv[key]
	if ok {
		return value
	}
	return fallback
}

func loadInt(fileEnv map[string]string, key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		value, ok = fileEnv[key]
	}
	if !ok {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func loadDuration(fileEnv map[string]string, key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		value, ok = fileEnv[key]
	}
	if !ok {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
