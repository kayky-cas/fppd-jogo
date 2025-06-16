package main

import (
	"errors"
	"fppd_jogo/pkg/jogo"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
)

type Args struct {
	A, B int
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
	eventosLocker EventosLocker
	jogoLocker    JogoLocker
}

type EntrouArgs struct{}

func (s *State) Entrar(args *EntrouArgs, reply *jogo.Jogo) error {
	jogador := jogo.Jogador{
		X: 10,
		Y: 14,
	}

	s.jogoLocker.Lock()

	jogador.ID = len(s.jogoLocker.jogo.Jogadores)

	s.jogoLocker.jogo.JogadorID = jogador.ID
	s.jogoLocker.jogo.Jogadores = append(s.jogoLocker.jogo.Jogadores, jogador)
	*reply = s.jogoLocker.jogo

	s.jogoLocker.Unlock()

	s.eventosLocker.Lock()
	s.eventosLocker.eventos = append(s.eventosLocker.eventos, jogo.NovoEntrarJogadorEvento(jogador.ID, jogador.X, jogador.Y))
	s.eventosLocker.Unlock()

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

func (s *State) CheckEventos(args *int, reply *[]jogo.Evento) error {
	if args == nil {
		return errors.New("Não foi enviado nenhum id de evento")
	}

	s.eventosLocker.Lock()
	if len(s.eventosLocker.eventos) < *args {
		return errors.New("Número de evento inválido")
	}

	*reply = s.eventosLocker.eventos[*args+1:]
	s.eventosLocker.Unlock()

	return nil
}

func main() {
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	state := State{
		eventosLocker: EventosLocker{eventos: make([]jogo.Evento, 0)},
		jogoLocker: JogoLocker{
			jogo: jogo.JogoNovo(),
		},
	}

	if err := jogo.JogoCarregarMapa(mapaFile, &state.jogoLocker.jogo); err != nil {
		panic(err)
	}

	rpc.Register(&state)

	listener, err := net.Listen("tcp", ":3000")

	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	defer listener.Close()

	rpc.Accept(listener)
}
