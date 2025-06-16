package jogo

import (
	"log"
	"net/rpc"
)

type Evento struct {
	ID      int            `json:"id"`
	Tipo    TipoEvento     `json:"tipo"`
	Payload map[string]any `json:"payload"`
}

type TipoEvento int

const (
	TipoEventoNovoJogador TipoEvento = iota
	TipoEventoMoveJogador
)

func NovoMoveJogadorEvento(id, x, y int) Evento {
	payload := make(map[string]any)

	payload["id"] = id
	payload["x"] = x
	payload["y"] = y

	return Evento{
		-1,
		TipoEventoMoveJogador,
		payload,
	}
}

func NovoEntrarJogadorEvento(id, x, y int) Evento {
	payload := make(map[string]any)

	payload["id"] = id
	payload["x"] = x
	payload["y"] = y

	return Evento{
		-1,
		TipoEventoNovoJogador,
		payload,
	}
}

func HandleEvento(j *Jogo, evento *Evento) {
	j.UltimoEvento = evento.ID
	switch evento.Tipo {
	case TipoEventoMoveJogador:
		id := evento.Payload["id"].(int)
		x := evento.Payload["x"].(int)
		y := evento.Payload["y"].(int)

		HandleMoveJogador(j, id, x, y)
	case TipoEventoNovoJogador:
		id := evento.Payload["id"].(int)
		x := evento.Payload["x"].(int)
		y := evento.Payload["y"].(int)

		HandleNovoJogador(j, id, x, y)
	}
}

func HandleNovoJogador(j *Jogo, id, x, y int) {
	if len(j.Jogadores) > id {
		return
	}

	j.Jogadores = append(j.Jogadores, Jogador{
		id,
		x,
		y,
		true,
	})
}

func HandleMoveJogador(jogo *Jogo, id, x, y int) {
	jogo.Jogadores[id].X = x
	jogo.Jogadores[id].Y = y
}

func EnviarEvento(client *rpc.Client, evento Evento) {
	var id int
	err := client.Call("State.Evento", &evento, &id)

	if err != nil {
		log.Print(err)
	}
}
