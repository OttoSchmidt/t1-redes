package main

import (
	"fmt"
	"os"
	"syscall"

	debug "pacman-redes/lib/debug"
	pacman "pacman-redes/lib/pacman"
	rawsockets "pacman-redes/lib/rawSockets"
)

func main() {
	// interface padrao: loopback. para usar outra interface,
	// passe o nome como argumento,como: eth0, enp3s0...
	ifaceName := "lo"
	csvMap := "./files/ufpr.csv"
	if len(os.Args) > 1 {
		ifaceName = os.Args[1]

		if len(os.Args) > 2 {
			csvMap = os.Args[2]
		}
	}

	// iniciar socket
	err := rawsockets.CreateSocket(ifaceName)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(rawsockets.ServerState.Sock)

	// iniciar jogo
	var gs pacman.GameState
	err = gs.ReadMapCsv(csvMap)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	fmt.Println("Servidor iniciado")

	// enviar mapa
	content := gs.GameMap.ToBytes()
	err = rawsockets.SendContent(content, rawsockets.Init)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 256)

	// loop do jogo
	gameRunning := true
	for gameRunning {
		// escutar por movimento
		content, packetType, err := rawsockets.ReceiveContent(buf)
		if err != nil {
			fmt.Printf("erro ao receber pacote do cliente: %s\n", err.Error())
			continue
		}

		debug.WriteDebug("conteudo recebido do cliente: %s\n", content)

		gs.IncrementRound()

		switch packetType {
		case rawsockets.MoveUp, rawsockets.MoveDown, 
			rawsockets.MoveLeft, rawsockets.MoveRight:
				err = gs.MovePlayer(packetType)
				if err != nil {
					fmt.Printf("erro ao movimentar player: %s\n", err.Error())
				}

				if gs.CoinsCollected == 6 {
					gameRunning = false
					break
				} 

				// enviar novo mapa
				content := gs.GameMap.ToBytes()
				err = rawsockets.SendContent(content, rawsockets.Visualize)
				if err != nil {
					fmt.Printf("erro ao enviar novo mapa: %s\n", err.Error())
				}
		case rawsockets.EndConn:
			gameRunning = false
		}
	}

	err = rawsockets.SendContent(nil, rawsockets.EndConn)
	if err != nil {
		fmt.Printf("erro ao enviar msg de finalizacao: %s\n", err.Error())
	}

	fmt.Println("Servidor finalizado")
}
