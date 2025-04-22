// main.go - Loop principal do jogo
package main

import (
	"math"
	"os"
	"time"
)

const (
	Cima     = 1
	Baixo    = 2
	Esquerda = 4
	Direita  = 8
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
			jogo.mutex.Lock()
			interfaceDesenharJogo(&jogo)
			jogo.mutex.Unlock()
			break
		case EventoDerrota:
			return
		}
	}
}

func prepararInimigos(jogo *Jogo, eventCh chan Evento) {
	for j, linha := range jogo.Mapa {
		for i, elemento := range linha {
			if elemento.simbolo == '☠' {
				go seguirPlayer(jogo, i, j, eventCh)
			}
		}
	}
}

func seguirPlayer(jogo *Jogo, i, j int, eventCh chan Evento) {
	var direcoes int

	for {
		time.Sleep(1000 * time.Millisecond)

		direcoes = 0

		for {
			if direcoes == Cima|Baixo|Direita|Esquerda {
				return
			}

			jogo.mutex.Lock()
			playerX, playerY := jogo.PosX, jogo.PosY

			jogo.mutex.Unlock()
			nextPosX, nextPosY := proximaAcao(playerX, playerY, i, j, &direcoes)
			jogo.mutex.Lock()

			if jogo.PosX == nextPosX && jogo.PosY == nextPosY {
				eventCh <- EventoDerrota
				return
			}

			elemento := jogo.Mapa[nextPosY][nextPosX]

			if elemento.tangivel {
				jogo.mutex.Unlock()
				continue
			}

			jogo.Mapa[nextPosY][nextPosX] = jogo.Mapa[j][i]
			jogo.Mapa[j][i] = elemento

			jogo.mutex.Unlock()

			i, j = nextPosX, nextPosY

			eventCh <- EventoDesenha

			break
		}
	}
}

func proximaAcao(playerX, playerY, inimigoX, inimigoY int, direcoes *int) (int, int) {
	dx := playerX - inimigoX
	dy := playerY - inimigoY

	absDx := math.Abs(float64(dx))
	absDy := math.Abs(float64(dy))

	moveX := func() bool {
		if dx > 0 && *direcoes&Direita == 0 {
			*direcoes |= Direita
			inimigoX += 1
			return true
		} else if dx < 0 && *direcoes&Esquerda == 0 {
			*direcoes |= Esquerda
			inimigoX -= 1
			return true
		}
		return false
	}

	moveY := func() bool {
		if dy > 0 && *direcoes&Baixo == 0 {
			*direcoes |= Baixo
			inimigoY += 1
			return true
		} else if dy < 0 && *direcoes&Cima == 0 {
			*direcoes |= Cima
			inimigoY -= 1
			return true
		}
		return false
	}

	if absDx > absDy {
		if !moveX() {
			moveY()
		}
	} else {
		if !moveY() {
			moveX()
		}
	}

	return inimigoX, inimigoY
}

func moveJogador(jogo *Jogo, eventCh chan<- Evento) {
	for {
		evento := interfaceLerEventoTeclado()
		jogo.mutex.Lock()
		if continuar := personagemExecutarAcao(evento, jogo); !continuar {
			eventCh <- EventoSair
		}
		jogo.mutex.Unlock()
		eventCh <- EventoDesenha
	}
}
