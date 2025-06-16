package main

import (
	"fppd_jogo/pkg/jogo"
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:3000")

	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	defer client.Close()

	var result []jogo.Evento
	i := 0
	err = client.Call("State.CheckEventos", &i, &result)

	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	log.Printf("RESULT: %v", result)
}
