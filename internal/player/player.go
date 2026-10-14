// Package player guarda a identidade do jogador e a regra do nome. Não
// conhece UI nem transporte: a mesma regra vale no jogo local e no servidor.
package player

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Limites do nome. O teto existe porque o placar precisa alinhar.
const (
	NameMinLen = 3
	NameMaxLen = 16
)

var (
	// ErrNameInvalid é formato recusado.
	ErrNameInvalid = errors.New("nome inválido")
	// ErrNameTaken é nome já reivindicado por outra identidade.
	ErrNameTaken = errors.New("nome já está em uso")
)

// Um espaço interno no máximo, nunca nas pontas nem dobrado: o grupo opcional
// expressa exatamente isso.
var nameRE = regexp.MustCompile(`^[a-z0-9_-]+( [a-z0-9_-]+)?$`)

// CleanName devolve o nome canônico ou ErrNameInvalid. É o único caminho para
// um nome entrar no sistema, então nenhum chamador precisa lembrar de validar.
func CleanName(s string) (string, error) {
	// Caractere de controle é recusado ANTES do trim. Ao contrário: o
	// TrimSpace removeria o "\n" de "afonso\n" e um nome com controle
	// embutido passaria a válido.
	if strings.ContainsFunc(s, unicode.IsControl) {
		return "", fmt.Errorf("%w: caractere de controle", ErrNameInvalid)
	}

	s = strings.TrimSpace(s)
	if len(s) < NameMinLen || len(s) > NameMaxLen {
		return "", fmt.Errorf("%w: precisa de %d a %d caracteres", ErrNameInvalid, NameMinLen, NameMaxLen)
	}
	if !nameRE.MatchString(s) {
		return "", fmt.Errorf("%w: %s", ErrNameInvalid, NameRule())
	}
	if !strings.ContainsFunc(s, isAlnum) {
		return "", fmt.Errorf("%w: precisa de pelo menos uma letra ou número", ErrNameInvalid)
	}
	return s, nil
}

// NameRule descreve a regra para mostrar na tela, para ninguém descobrir o
// formato por tentativa e erro.
func NameRule() string {
	return fmt.Sprintf("%d a %d: minúsculas, números, - e _, no máximo um espaço", NameMinLen, NameMaxLen)
}

func isAlnum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// Kind diz de onde a identidade veio, o que determina se ela persiste.
type Kind string

const (
	KindPubKey Kind = "pubkey" // chave SSH: estável entre sessões
	KindLocal  Kind = "local"  // jogo local, usuário do SO
	KindGuest  Kind = "guest"  // SSH sem chave: nome por sessão
)

// Identity é quem está jogando nesta sessão.
type Identity struct {
	ID   string
	Kind Kind
	Name string // vazio quando a pessoa ainda não se apresentou
}

// Persistent informa se vale gravar esta identidade.
func (i Identity) Persistent() bool { return i.Kind != KindGuest }

// Player é a identidade gravada.
type Player struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// Store persiste identidades. Interface pequena de propósito: trocar JSON por
// banco depois é implementar três métodos.
type Store interface {
	Get(ctx context.Context, id string) (Player, bool, error)
	// Claim grava id→nome se o nome estiver livre, senão devolve ErrNameTaken.
	// Uma chamada só em vez de Existe+Grava porque é menos código; colisão
	// simultânea entre duas sessões está postergada no spec.
	Claim(ctx context.Context, id, name string) (Player, error)
	// NameTaken diz se o nome já pertence a outra identidade. Serve ao
	// convidado, que precisa de nome livre mas não reserva nada.
	NameTaken(ctx context.Context, name string) (bool, error)
	Touch(ctx context.Context, id string) error
}
