// jo()go.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package jogo

import (
	"bufio"
	"net/rpc"
	"os"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	Simbolo  rune `json:"simbolo,omitempty"`
	Cor      Cor  `json:"cor,omitempty"`
	CorFundo Cor  `json:"cor_fundo,omitempty"`
	Tangivel bool `json:"tangivel,omitempty"` // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa         [][]Elemento `json:"mapa,omitempty"`   // grade 2D representando o mapa
	StatusMsg    string       `json:"status,omitempty"` // mensagem para a barra de status
	JogadorID    int          `json:"jogador,omitempty"`
	Jogadores    []Jogador    `json:"jogadores,omitempty"`
	UltimoEvento int          `json:"ultimo_evento,omitempty"`
	rpcClient    *rpc.Client
}

func (j *Jogo) SetRPCClient(client *rpc.Client) {
	j.rpcClient = client
}

func (j *Jogo) Jogador() *Jogador {
	return &j.Jogadores[j.JogadorID]
}

// Elementos visuais do jogo
var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func JogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	jogadores := make([]Jogador, 0, 10)

	return Jogo{Jogadores: jogadores, JogadorID: -1, UltimoEvento: -1}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func JogoCarregarMapa(nome string, jogo *Jogo) error {
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
		for _, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.Simbolo:
				e = Parede
			case Inimigo.Simbolo:
				e = Inimigo
			case Vegetacao.Simbolo:
				e = Vegetacao
				// case Personagem.simbolo:
				// 	jogo.jogador.PosX, jogo.jogador.PosY = x, y // registra a posição inicial do personagem
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
	if jogo.Mapa[y][x].Tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	// nx, ny := x+dx, y+dy

	// jogo.Jogador().X = nx
	// jogo.Jogador().Y = ny
}
