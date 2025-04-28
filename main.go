// main.go - Loop principal do jogo
package main

import (
	"fmt"
	"os"
	"time"
)

type Evento int

const (
	EventoSair Evento = iota
	EventoDesenha
	EventoDerrota
	EventoInimigoNasceu
	EventoInimigoMorreu
)

func main() {
	// Inicializa a interface (termbox)
	interfaceIniciar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// Inicializa o jogo
	jogo := jogoNovo(time.Second * 10)
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Desenha o estado inicial do jogo
	interfaceDesenharJogo(&jogo)

	eventCh := make(chan Evento)

	go moveJogador(&jogo, eventCh)
	go prepararInimigos(&jogo, eventCh)
	go timer(&jogo, eventCh)

	inimigoCont := 0

	// Loop principal de entrada
	for event := range eventCh {
		switch event {
		case EventoSair:
			interfaceFinalizar()
			return
		case EventoDesenha:
			jogo.Mutex.Lock()
			interfaceDesenharJogo(&jogo)
			jogo.Mutex.Unlock()
			break
		case EventoInimigoNasceu:
			inimigoCont += 1
			break
		case EventoInimigoMorreu:
			inimigoCont -= 1

			if inimigoCont == 0 {
				interfaceFinalizar()
				fmt.Println("Parabéns você ganhou!")
				return
			}

			break
		case EventoDerrota:
			interfaceFinalizar()
			fmt.Println("Por favor melhore...")
			return
		}
	}
}
