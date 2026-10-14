package score

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// JSONStore guarda o placar num arquivo JSON.
type JSONStore struct {
	path string
	mu   sync.Mutex
}

// NewJSONStore devolve um store apontando para path, que não precisa existir.
func NewJSONStore(path string) *JSONStore { return &JSONStore{path: path} }

type scoresFile struct {
	Entries []Entry `json:"entries"`
}

// Top devolve as n melhores entradas. Empate desempata pela mais antiga, para
// a ordem não dançar entre leituras.
func (s *JSONStore) Top(_ context.Context, n int) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return nil, err
	}
	entries := data.Entries
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Points != entries[j].Points {
			return entries[i].Points > entries[j].Points
		}
		return entries[i].At.Before(entries[j].At)
	})
	if n > 0 && len(entries) > n {
		entries = entries[:n]
	}
	return entries, nil
}

// Save acrescenta uma partida ao placar.
func (s *JSONStore) Save(_ context.Context, e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return err
	}
	data.Entries = append(data.Entries, e)
	return s.save(data)
}

func (s *JSONStore) load() (scoresFile, error) {
	var data scoresFile
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
	return data, nil
}

// save grava em temporário e renomeia, pelo mesmo motivo do store de
// identidades: crash no meio não deixa JSON truncado.
func (s *JSONStore) save(data scoresFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("criar diretório de %s: %w", s.path, err)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("codificar placar: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".scores-*.json")
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
