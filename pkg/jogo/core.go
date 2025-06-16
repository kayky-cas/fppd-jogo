package jogo

type Core struct {
	jogo  *Jogo
	cache any
}

type Jogador struct {
	ID     int  `json:"id,omitempty"`
	X      int  `json:"x,omitempty"`
	Y      int  `json:"y,omitempty"`
	Online bool `json:"online,omitempty"`
}
