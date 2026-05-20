package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigMergesEnvFilesFromMonorepoRoot(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, ".git"))
	mustMkdirAll(t, filepath.Join(repoRoot, "reference", "workflow-backend-skeleton"))

	mustWriteFile(t, filepath.Join(repoRoot, ".env"), ""+
		"HTTP_ADDR=:8080\n"+
		"DB_DSN=base-dsn\n"+
		"DB_MAX_OPEN_CONNS=20\n"+
		"DB_MAX_IDLE_CONNS=10\n"+
		"DB_CONN_MAX_LIFETIME=1h\n")
	mustWriteFile(t, filepath.Join(repoRoot, ".env.development"), ""+
		"HTTP_ADDR=:8081\n"+
		"DB_DSN=dev-dsn\n"+
		"DB_MAX_OPEN_CONNS=25\n")
	mustWriteFile(t, filepath.Join(repoRoot, ".env.local"), ""+
		"DB_DSN=local-dsn\n"+
		"DB_MAX_IDLE_CONNS=12\n"+
		"DB_CONN_MAX_LIFETIME=90m\n")

	unsetEnv(t, "NODE_ENV")
	unsetEnv(t, "HTTP_ADDR")
	unsetEnv(t, "DB_DSN")
	unsetEnv(t, "DB_MAX_OPEN_CONNS")
	unsetEnv(t, "DB_MAX_IDLE_CONNS")
	unsetEnv(t, "DB_CONN_MAX_LIFETIME")

	withWorkingDir(t, filepath.Join(repoRoot, "reference", "workflow-backend-skeleton"), func() {
		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if cfg.HTTP.Addr != ":8081" {
			t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, ":8081")
		}
		if cfg.DB.DSN != "local-dsn" {
			t.Fatalf("DB.DSN = %q, want %q", cfg.DB.DSN, "local-dsn")
		}
		if cfg.DB.MaxOpenConns != 25 {
			t.Fatalf("DB.MaxOpenConns = %d, want 25", cfg.DB.MaxOpenConns)
		}
		if cfg.DB.MaxIdleConns != 12 {
			t.Fatalf("DB.MaxIdleConns = %d, want 12", cfg.DB.MaxIdleConns)
		}
		if cfg.DB.ConnMaxLifetime != 90*time.Minute {
			t.Fatalf("DB.ConnMaxLifetime = %s, want %s", cfg.DB.ConnMaxLifetime, 90*time.Minute)
		}
	})
}

func TestLoadConfigSystemEnvOverridesFileEnv(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, ".git"))

	mustWriteFile(t, filepath.Join(repoRoot, ".env"), ""+
		"HTTP_ADDR=:8080\n"+
		"DB_DSN=base-dsn\n")
	mustWriteFile(t, filepath.Join(repoRoot, ".env.production"), ""+
		"HTTP_ADDR=:8082\n"+
		"DB_DSN=prod-dsn\n")
	mustWriteFile(t, filepath.Join(repoRoot, ".env.local"), ""+
		"HTTP_ADDR=:8083\n"+
		"DB_DSN=local-dsn\n")

	t.Setenv("NODE_ENV", "production")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DB_DSN", "system-dsn")

	withWorkingDir(t, repoRoot, func() {
		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if cfg.HTTP.Addr != ":9090" {
			t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, ":9090")
		}
		if cfg.DB.DSN != "system-dsn" {
			t.Fatalf("DB.DSN = %q, want %q", cfg.DB.DSN, "system-dsn")
		}
	})
}

func TestLoadConfigReturnsErrorForInvalidEnvFile(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, ".git"))
	mustWriteFile(t, filepath.Join(repoRoot, ".env.local"), "BROKEN_LINE\n")

	unsetEnv(t, "NODE_ENV")

	withWorkingDir(t, repoRoot, func() {
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("LoadConfig() error = nil, want error")
		}
	})
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q) error = %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldDir)
	})

	fn()
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) error = %v", key, err)
	}

	t.Cleanup(func() {
		if !existed {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, oldValue)
	})
}
