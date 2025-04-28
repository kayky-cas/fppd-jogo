// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"math"
	"os"
	"sync"
	"time"
)

type Direcao int

const (
	Cima     Direcao = 1
	Baixo            = 2
	Esquerda         = 4
	Direita          = 8
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool // Indica se o elemento bloqueia passagem
}

type Bala struct {
	X          int
	Y          int
	Direcao    Direcao
	Habilitada bool
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Balas          []Bala
	Mapa           [][]Elemento // grade 2D representando o mapa
	PosX, PosY     int          // posição atual do personagem
	UltimoVisitado Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg      string       // mensagem para a barra de status
	Mutex          sync.Mutex
	Direcao        Direcao
}

// Elementos visuais do jogo
var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, true}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{UltimoVisitado: Vazio}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Inimigo
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	// Verifica se a coordenada Y está dentro dos limites verticais do mapa
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	// Verifica se a coordenada X está dentro dos limites horizontais do mapa
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado   // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx] // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento            // move o elemento
}

func atirarBala(jogo *Jogo, eventCh chan Evento) {
	jogo.Mutex.Lock()
	x, y, direcao := jogo.PosX, jogo.PosY, jogo.Direcao

	jogo.Balas = append(jogo.Balas, Bala{x, y, direcao, true})
	bala := &jogo.Balas[len(jogo.Balas)-1]
	jogo.Mutex.Unlock()

	for {
		time.Sleep(100 * time.Millisecond)
		jogo.Mutex.Lock()

		switch bala.Direcao {
		case Cima:
			bala.Y -= 1
		case Baixo:
			bala.Y += 1
		case Esquerda:
			bala.X -= 1
		case Direita:
			bala.X += 1
		}

		if jogo.Mapa[bala.Y][bala.X].tangivel {
			if jogo.Mapa[bala.Y][bala.X].simbolo == '☠' {
				jogo.Mapa[bala.Y][bala.X] = Vazio
				eventCh <- EventoInimigoMorreu
			}

			bala.Habilitada = false
			jogo.Mutex.Unlock()
			eventCh <- EventoDesenha
			return
		}

		jogo.Mutex.Unlock()

		eventCh <- EventoDesenha
	}
}

func prepararInimigos(jogo *Jogo, eventCh chan Evento) {
	for j, linha := range jogo.Mapa {
		for i, elemento := range linha {
			if elemento.simbolo == '☠' {
				eventCh <- EventoInimigoNasceu
				go seguirPlayer(jogo, i, j, eventCh)
			}
		}
	}
}

func seguirPlayer(jogo *Jogo, i, j int, eventCh chan Evento) {
	var direcoes Direcao

	for {
		time.Sleep(1000 * time.Millisecond)

		direcoes = 0

		for {
			if direcoes == Cima|Baixo|Direita|Esquerda {
				return
			}

			jogo.Mutex.Lock()
			if jogo.Mapa[j][i].simbolo != '☠' {
				jogo.Mutex.Unlock()
				return
			}
			playerX, playerY := jogo.PosX, jogo.PosY
			jogo.Mutex.Unlock()

			nextPosX, nextPosY := proximaAcao(playerX, playerY, i, j, &direcoes)

			jogo.Mutex.Lock()

			if jogo.Mapa[j][i].simbolo != '☠' {
				jogo.Mutex.Unlock()
				return
			}

			if jogo.PosX == nextPosX && jogo.PosY == nextPosY {
				eventCh <- EventoDerrota
				return
			}

			elemento := jogo.Mapa[nextPosY][nextPosX]

			if elemento.tangivel {
				jogo.Mutex.Unlock()
				continue
			}

			jogo.Mapa[nextPosY][nextPosX] = jogo.Mapa[j][i]
			jogo.Mapa[j][i] = elemento

			jogo.Mutex.Unlock()

			i, j = nextPosX, nextPosY

			eventCh <- EventoDesenha

			break
		}
	}
}

func proximaAcao(playerX, playerY, inimigoX, inimigoY int, direcoes *Direcao) (int, int) {
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

func moveJogador(jogo *Jogo, eventCh chan Evento) {
	for {
		evento := interfaceLerEventoTeclado()
		jogo.Mutex.Lock()
		if continuar := personagemExecutarAcao(evento, jogo, eventCh); !continuar {
			eventCh <- EventoSair
		}
		jogo.Mutex.Unlock()
		eventCh <- EventoDesenha
	}
}
