package ui

// Tamanhos de tela. O default entra quando o cliente não diz o tamanho, o piso
// evita layout impossível numa janela minúscula.
const (
	DefaultWidth  = 80
	DefaultHeight = 24
	MinWidth      = 40
	MinHeight     = 10
)

// Size normaliza o tamanho anunciado pelo cliente. Um PTY pode chegar 0x0 —
// visto na prática — e aí o Bubble Tea não desenharia nada.
func Size(w, h int) (int, int) {
	if w <= 0 {
		w = DefaultWidth
	}
	if h <= 0 {
		h = DefaultHeight
	}
	if w < MinWidth {
		w = MinWidth
	}
	if h < MinHeight {
		h = MinHeight
	}
	return w, h
}
