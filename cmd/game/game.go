// main.go - Loop principal do jogo
package main

import (
	"fmt"
	jogo "fppd_jogo/pkg/jogo"
	"log"
	"net/rpc"
	"time"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:3000")

	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	defer client.Close()

	var j jogo.Jogo
	err = client.Call("State.Entrar", struct{}{}, &j)

	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	j.SetRPCClient(client)

	// Inicializa a interface (termbox)
	jogo.InterfaceIniciar()
	defer jogo.InterfaceFinalizar()

	// Desenha o estado inicial do jogo
	jogo.InterfaceDesenharJogo(&j)

	go func() {
		var checkArgs jogo.CheckArgs
		for {
			select {
			case <-time.NewTicker(time.Second / 60).C:
				var eventos []jogo.Evento
				checkArgs = jogo.CheckArgs{ID: j.Jogador().ID, UltimoEvento: j.UltimoEvento}
				err := client.Call("State.CheckEventos", &checkArgs, &eventos)

				if err != nil {
					continue
				}

				if len(eventos) > 0 {
					j.StatusMsg = fmt.Sprint(eventos)
					for _, evento := range eventos {
						jogo.HandleEvento(&j, &evento)
					}
					jogo.InterfaceDesenharJogo(&j)
				}
			}
		}
	}()

	// Loop principal de entrada
	for {
		evento := jogo.InterfaceLerEventoTeclado()
		if continuar := jogo.PersonagemExecutarAcao(evento, &j); !continuar {
			break
		}
	}
}
