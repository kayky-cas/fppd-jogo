package main

import (
	"errors"
	"fppd_jogo/pkg/jogo"
	"log"
	"math"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type JogadorStatus struct {
	UltimoCheck  int64
	Desconectado bool
}

type JogadoresStatusLocker struct {
	sync.Mutex
	jogadoresStatus []JogadorStatus
}

type EventosLocker struct {
	sync.Mutex
	eventos []jogo.Evento
}

type JogoLocker struct {
	sync.Mutex
	jogo jogo.Jogo
}

type State struct {
	eventosLocker         EventosLocker
	jogoLocker            JogoLocker
	jogadoresStatusLocker JogadoresStatusLocker
}

type EntrouArgs struct{}

type Direcao int

const (
	Cima     Direcao = 1
	Baixo            = 2
	Esquerda         = 4
	Direita          = 8
)

func seguirPlayer(j *jogo.Jogo, x, y int, playerX, playerY int) (int, int, bool) {
	var direcoes Direcao

	direcoes = 0

	for {
		if direcoes == Cima|Baixo|Direita|Esquerda {
			return -1, -1, false
		}

		nextPosX, nextPosY := proximaAcao(playerX, playerY, x, y, &direcoes)

		if j.Mapa[nextPosY][nextPosX].Tangivel {
			continue
		}

		return nextPosX, nextPosY, true
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

func jogadorMaisProximo(j *jogo.Jogo, inimigo *jogo.Vilao) int {
	var distMaisProxima int = 1000
	var idMaisProximo int

	for _, jogador := range j.Jogadores {
		dist := int(math.Abs(float64(jogador.X-inimigo.X)) + math.Abs(float64(jogador.Y-inimigo.Y)))

		if dist < distMaisProxima {
			distMaisProxima = dist
			idMaisProximo = jogador.ID
		}
	}

	return idMaisProximo
}

func (s *State) MoverInimigos() {
	for range time.NewTicker(time.Second).C {
		s.jogoLocker.Lock()
		if len(s.jogoLocker.jogo.Jogadores) < 1 {
			s.jogoLocker.Unlock()
			continue
		}
		for i := range s.jogoLocker.jogo.Inimigos {
			inimigo := &s.jogoLocker.jogo.Inimigos[i]
			jogadorID := jogadorMaisProximo(&s.jogoLocker.jogo, inimigo)

			x, y, ok := seguirPlayer(&s.jogoLocker.jogo, inimigo.X, inimigo.Y, s.jogoLocker.jogo.Jogadores[jogadorID].X, s.jogoLocker.jogo.Jogadores[jogadorID].Y)

			if !ok {
				continue
			}

			inimigo.X = x
			inimigo.Y = y

			evento := jogo.NovoMoveInimigoEvento(s.jogoLocker.jogo.Inimigos[i].ID, s.jogoLocker.jogo.Inimigos[i].X, s.jogoLocker.jogo.Inimigos[i].Y)
			s.eventosLocker.Lock()
			evento.ID = len(s.eventosLocker.eventos)
			s.eventosLocker.eventos = append(s.eventosLocker.eventos, evento)
			s.eventosLocker.Unlock()
		}
		s.jogoLocker.Unlock()
	}
}

func (s *State) CheckOffline() {
	now := time.Now().Unix()
	for range time.NewTicker(time.Second * 2).C {
		s.jogadoresStatusLocker.Lock()
		for i := range s.jogadoresStatusLocker.jogadoresStatus {
			status := &s.jogadoresStatusLocker.jogadoresStatus[i]

			if status.Desconectado {
				continue
			}

			if status.UltimoCheck < now {
				status.Desconectado = true

				evento := jogo.NovoDesconectaJogadorEvento(i)
				s.eventosLocker.Lock()
				evento.ID = len(s.eventosLocker.eventos)
				s.eventosLocker.eventos = append(s.eventosLocker.eventos, evento)
				s.eventosLocker.Unlock()
			}
		}
		s.jogadoresStatusLocker.Unlock()
	}
}

func (s *State) Entrar(args *EntrouArgs, reply *jogo.Jogo) error {
	jogador := jogo.Jogador{
		X:      10,
		Y:      14,
		Online: true,
	}

	s.jogoLocker.Lock()
	jogador.ID = len(s.jogoLocker.jogo.Jogadores)

	s.jogoLocker.jogo.JogadorID = jogador.ID
	s.jogoLocker.jogo.Jogadores = append(s.jogoLocker.jogo.Jogadores, jogador)

	s.jogoLocker.Unlock()

	*reply = s.jogoLocker.jogo

	evento := jogo.NovoEntrarJogadorEvento(jogador.ID, jogador.X, jogador.Y)
	s.eventosLocker.Lock()
	evento.ID = len(s.eventosLocker.eventos)
	s.eventosLocker.eventos = append(s.eventosLocker.eventos, evento)
	s.eventosLocker.Unlock()

	s.jogadoresStatusLocker.Lock()
	s.jogadoresStatusLocker.jogadoresStatus = append(s.jogadoresStatusLocker.jogadoresStatus, JogadorStatus{time.Now().Unix(), false})
	s.jogadoresStatusLocker.Unlock()
	return nil
}

func (s *State) Evento(args *jogo.Evento, reply *int) error {
	if args == nil {
		return errors.New("Não foi enviado nenhum evento")
	}

	s.eventosLocker.Lock()
	id := len(s.eventosLocker.eventos)
	args.ID = id
	s.eventosLocker.eventos = append(s.eventosLocker.eventos, *args)
	s.eventosLocker.Unlock()

	s.jogoLocker.Lock()
	jogo.HandleEvento(&s.jogoLocker.jogo, args)
	s.jogoLocker.Unlock()

	*reply = id

	return nil
}

func (s *State) CheckEventos(args *jogo.CheckArgs, reply *[]jogo.Evento) error {
	if args == nil {
		return errors.New("Não foi enviado nenhum id de evento")
	}

	s.jogadoresStatusLocker.Lock()
	s.jogadoresStatusLocker.jogadoresStatus[args.ID].UltimoCheck = time.Now().Unix()
	s.jogadoresStatusLocker.Unlock()

	s.eventosLocker.Lock()
	if len(s.eventosLocker.eventos) < args.UltimoEvento {
		return errors.New("Número de evento inválido")
	}

	*reply = s.eventosLocker.eventos[args.UltimoEvento+1:]
	s.eventosLocker.Unlock()

	return nil
}

func main() {
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	state := State{
		eventosLocker: EventosLocker{eventos: make([]jogo.Evento, 0, 1024)},
		jogoLocker: JogoLocker{
			jogo: jogo.JogoNovo(),
		},
		jogadoresStatusLocker: JogadoresStatusLocker{
			jogadoresStatus: make([]JogadorStatus, 10),
		},
	}

	if err := jogo.JogoCarregarMapa(mapaFile, &state.jogoLocker.jogo); err != nil {
		panic(err)
	}

	go state.MoverInimigos()
	// go state.CheckOffline()

	rpc.Register(&state)

	listener, err := net.Listen("tcp", ":3000")

	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	defer listener.Close()

	rpc.Accept(listener)
}
