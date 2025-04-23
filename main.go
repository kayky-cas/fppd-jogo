// main.go - Loop principal do jogo
package main

import (
	"os"
)

type Direcao int

const (
	Cima     Direcao = 1
	Baixo            = 2
	Esquerda         = 4
	Direita          = 8
)

type Evento int

const (
	EventoSair Evento = iota
	EventoDesenha
	EventoDerrota
)

func main() {
	// Inicializa a interface (termbox)
	interfaceIniciar()
	defer interfaceFinalizar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// Inicializa o jogo
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Desenha o estado inicial do jogo
	interfaceDesenharJogo(&jogo)

	eventCh := make(chan Evento)

	go moveJogador(&jogo, eventCh)
	prepararInimigos(&jogo, eventCh)

	// Loop principal de entrada
	for event := range eventCh {
		switch event {
		case EventoSair:
			return
		case EventoDesenha:
			jogo.Mutex.Lock()
			interfaceDesenharJogo(&jogo)
			jogo.Mutex.Unlock()
			break
		case EventoDerrota:
			return
		}
	}
}
