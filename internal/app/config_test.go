package app

import (
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("YATZEE_DATA_DIR", "")
	t.Setenv("YATZEE_SSH_HOST", "")
	t.Setenv("YATZEE_SSH_PORT", "")
	t.Setenv("YATZEE_SSH_HOST_KEY", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.SSHHost != "0.0.0.0" {
		t.Errorf("SSHHost = %q, quero 0.0.0.0", cfg.SSHHost)
	}
	if cfg.SSHPort != "2222" {
		t.Errorf("SSHPort = %q, quero 2222", cfg.SSHPort)
	}
	if !filepath.IsAbs(cfg.DataDir) {
		t.Errorf("DataDir = %q, quero caminho absoluto", cfg.DataDir)
	}
	if want := filepath.Join(cfg.DataDir, "ssh_host_ed25519"); cfg.SSHHostKey != want {
		t.Errorf("SSHHostKey = %q, quero %q", cfg.SSHHostKey, want)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("YATZEE_DATA_DIR", dir)
	t.Setenv("YATZEE_SSH_HOST", "127.0.0.1")
	t.Setenv("YATZEE_SSH_PORT", "3333")
	t.Setenv("YATZEE_SSH_HOST_KEY", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.DataDir != dir {
		t.Errorf("DataDir = %q, quero %q", cfg.DataDir, dir)
	}
	if cfg.SSHHost != "127.0.0.1" || cfg.SSHPort != "3333" {
		t.Errorf("host/porta = %q/%q", cfg.SSHHost, cfg.SSHPort)
	}
	if cfg.PlayersPath() != filepath.Join(dir, "players.json") {
		t.Errorf("PlayersPath = %q", cfg.PlayersPath())
	}
	if cfg.ScoresPath() != filepath.Join(dir, "scores.json") {
		t.Errorf("ScoresPath = %q", cfg.ScoresPath())
	}
}

// Um serviço sob systemd ou container tem CWD imprevisível: caminho relativo
// na env tem que virar absoluto.
func TestLoadConfigTornaDataDirAbsoluto(t *testing.T) {
	t.Setenv("YATZEE_DATA_DIR", "dados-relativos")
	t.Setenv("YATZEE_SSH_HOST_KEY", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !filepath.IsAbs(cfg.DataDir) {
		t.Errorf("DataDir = %q, quero absoluto", cfg.DataDir)
	}
}
