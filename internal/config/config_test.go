package config

import (
	"os"
	"testing"
)

func clearEnv(t *testing.T) {
	for _, k := range []string{"DB_HOST", "DB_NAME", "DB_USERNAME", "DB_PASSWORD"} {
		t.Setenv(k, "") // registers restore on cleanup
		os.Unsetenv(k)  // godotenv does not override variables that are already set
	}
}

func TestLoad_ReadsEnvFile(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	env := "DB_HOST=h\nDB_NAME=n\nDB_USERNAME=u\nDB_PASSWORD=p\n"
	if err := os.WriteFile(dir+"/.env", []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := Config{Host: "h", DBName: "n", Username: "u", Password: "p"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoad_MissingEnvFile(t *testing.T) {
	clearEnv(t)
	t.Chdir(t.TempDir())

	if _, err := Load(); err == nil {
		t.Fatal("expected an error when .env is missing")
	}
}
