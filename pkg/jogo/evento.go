package jogo

import (
	"log"
	"net/rpc"
)

type CheckArgs struct {
	ID           int `json:"id"`
	UltimoEvento int `json:"ultimoEvento"`
}

type Evento struct {
	ID      int            `json:"id"`
	Tipo    TipoEvento     `json:"tipo"`
	Payload map[string]any `json:"payload"`
}

type TipoEvento int

const (
	TipoEventoNovoJogador TipoEvento = iota
	TipoEventoMoveJogador
	TipoEventoMoveInimigo
	TipoEventoDesconectaJogador
)

func NovoDesconectaJogadorEvento(id int) Evento {
	payload := make(map[string]any)
	payload["id"] = id
	return Evento{
		-1,
		TipoEventoDesconectaJogador,
		payload,
	}
}

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

func NovoMoveInimigoEvento(id, x, y int) Evento {
	payload := make(map[string]any)

	payload["id"] = id
	payload["x"] = x
	payload["y"] = y

	return Evento{
		-1,
		TipoEventoMoveInimigo,
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
		break
	case TipoEventoNovoJogador:
		id := evento.Payload["id"].(int)
		x := evento.Payload["x"].(int)
		y := evento.Payload["y"].(int)

		HandleNovoJogador(j, id, x, y)
		break
	case TipoEventoMoveInimigo:
		id := evento.Payload["id"].(int)
		x := evento.Payload["x"].(int)
		y := evento.Payload["y"].(int)

		HandleMoveInimigo(j, id, x, y)
		break
	case TipoEventoDesconectaJogador:
		id := evento.Payload["id"].(int)

		HandleDesconectaJogador(j, id)
		break
	}
}

func HandleDesconectaJogador(j *Jogo, id int) {
	if len(j.Jogadores) <= id {
		return
	}

	j.Jogadores[id].Online = false
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

func HandleMoveInimigo(jogo *Jogo, id, x, y int) {
	jogo.Inimigos[id].X = x
	jogo.Inimigos[id].Y = y
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
