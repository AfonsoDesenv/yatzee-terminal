package player

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONStore guarda identidades num arquivo JSON. Serve ao jogo local e ao
// servidor com poucas dezenas de jogadores; trocar por banco é implementar a
// mesma interface.
type JSONStore struct {
	path string
	mu   sync.Mutex
}

// NewJSONStore devolve um store apontando para path, que não precisa existir.
func NewJSONStore(path string) *JSONStore { return &JSONStore{path: path} }

type playersFile struct {
	Players map[string]Player `json:"players"`
}

// Get busca uma identidade pelo ID.
func (s *JSONStore) Get(_ context.Context, id string) (Player, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return Player{}, false, err
	}
	p, ok := data.Players[id]
	return p, ok, nil
}

// Claim grava id→nome se o nome estiver livre. Reivindicar o próprio nome de
// novo é idempotente, porque reconectar não deve falhar.
func (s *JSONStore) Claim(_ context.Context, id, name string) (Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return Player{}, err
	}
	for otherID, p := range data.Players {
		if p.Name == name && otherID != id {
			return Player{}, fmt.Errorf("%w: %q", ErrNameTaken, name)
		}
	}

	now := time.Now().UTC()
	p, existed := data.Players[id]
	if !existed {
		p = Player{ID: id, FirstSeen: now}
	}
	p.Name = name
	p.LastSeen = now
	data.Players[id] = p

	if err := s.save(data); err != nil {
		return Player{}, err
	}
	return p, nil
}

// NameTaken diz se o nome já pertence a alguma identidade.
func (s *JSONStore) NameTaken(_ context.Context, name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return false, err
	}
	for _, p := range data.Players {
		if p.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// Touch marca a identidade como vista agora. Silencioso para ID desconhecido:
// um convidado não tem registro e isso não é erro.
func (s *JSONStore) Touch(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return err
	}
	p, ok := data.Players[id]
	if !ok {
		return nil
	}
	p.LastSeen = time.Now().UTC()
	data.Players[id] = p
	return s.save(data)
}

// load lê o arquivo. Ausente conta como vazio: primeiro boot não é erro.
func (s *JSONStore) load() (playersFile, error) {
	data := playersFile{Players: map[string]Player{}}
	raw, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return data, nil
	}
	if err != nil {
		return data, fmt.Errorf("ler %s: %w", s.path, err)
	}
	if len(raw) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, fmt.Errorf("decodificar %s: %w", s.path, err)
	}
	if data.Players == nil {
		data.Players = map[string]Player{}
	}
	return data, nil
}

// save grava em arquivo temporário e renomeia: um crash no meio da escrita
// deixa o arquivo anterior intacto em vez de um JSON truncado.
func (s *JSONStore) save(data playersFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("criar diretório de %s: %w", s.path, err)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("codificar identidades: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".players-*.json")
	if err != nil {
		return fmt.Errorf("criar temporário: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("permissão do temporário: %w", err)
	}
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("escrever temporário: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("fechar temporário: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("renomear para %s: %w", s.path, err)
	}
	return nil
}
