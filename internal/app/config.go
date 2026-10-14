// Package app monta a sessão de jogo e roteia as telas.
package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config reúne tudo que vem do ambiente. Nada aqui é lido de novo depois do
// boot: uma sessão recebe a config já resolvida.
type Config struct {
	SSHHost    string
	SSHPort    string
	DataDir    string
	SSHHostKey string
}

// LoadConfig lê a configuração do ambiente, aplicando os defaults do spec.
func LoadConfig() (Config, error) {
	dataDir := os.Getenv("YATZEE_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("descobrir home do usuário: %w", err)
		}
		dataDir = filepath.Join(home, ".local", "share", "yatzee")
	}
	// Caminho relativo é inútil num serviço: o CWD de um systemd ou container
	// não é o que o operador imagina.
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return Config{}, fmt.Errorf("resolver caminho absoluto de %q: %w", dataDir, err)
	}

	hostKey := os.Getenv("YATZEE_SSH_HOST_KEY")
	if hostKey == "" {
		hostKey = filepath.Join(dataDir, "ssh_host_ed25519")
	}

	return Config{
		SSHHost:    envOr("YATZEE_SSH_HOST", "0.0.0.0"),
		SSHPort:    envOr("YATZEE_SSH_PORT", "2222"),
		DataDir:    dataDir,
		SSHHostKey: hostKey,
	}, nil
}

// PlayersPath é o arquivo de identidades.
func (c Config) PlayersPath() string { return filepath.Join(c.DataDir, "players.json") }

// ScoresPath é o arquivo de placar.
func (c Config) ScoresPath() string { return filepath.Join(c.DataDir, "scores.json") }

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
